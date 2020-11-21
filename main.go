/*
file:   mp3taginfo
author: johnny
descr:  The tag tool reads metadata from media files (as supported by the tag library).
usage:  mp3taginfo -src=<directory/file>
*/
package main

import (
	/*	"encoding/json"  */
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/OrangeBox72/mp3tagInfo/version"
	"github.com/dhowden/tag"
)

// ===== FUNCTIONS =====================================================
func usage() {
	fmt.Printf("%v - Version: %v\n", os.Args[0], version.BuildVersion)
  fmt.Fprintf(os.Stderr, "\n")
	flag.PrintDefaults()
	os.Exit(0)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type songCommonInfo struct {
	format     string
	idtype     string
	genre      string
	year       int
	album      string
	disc       int
	discCount  int
	trackCount int
	composer   string
	comment    string
	picture    string
	picSize    int
}

// ===== MAIN ==========================================================
func main() {
	// Vars ------------------------
	var filPtr *os.File
	var songMetadata tag.Metadata
//	var songMetadataCtrlVar tag.Metadata
  var songMetadataCtrlVarPrinted bool
	var songCtrlVar songCommonInfo
	var song songCommonInfo
	//	var pic1 *tag.Picture
	var pic *tag.Picture
	var err error
	var files []string
	var file string

  // commandline vars
	var allPtr *bool
	var all bool
	var ignoreCommentsPtr *bool
	var includeComposerPtr *bool
	var sourcePtr *string
	var source string
	var usagePtr *bool
  var versionPtr *bool
	var ignoreComments bool
	var includeComposer bool


  var errCode string
	var songTrack int
	var x int
	var titlePrinted bool
  var minimumPicSize int

	titlePrinted = false
  minimumPicSize = 4500
	// ----- Passed Args ----------
	allPtr = flag.Bool("all", false, "Prints info on ALL files regardless of Tag irregularities")
	ignoreCommentsPtr = flag.Bool("ignore-comments", false, "Ignores *comment* fields in comparisons")
	includeComposerPtr = flag.Bool("include-composer", false, "includes composer fields in comparisons")
	sourcePtr = flag.String("src", "", "<REQUIRED> Source of MP3's to parse for Tags")
	usagePtr = flag.Bool("usage", false, "This message")
  versionPtr = flag.Bool("version", false, "Version info")
	flag.Parse()

	all = *allPtr
	ignoreComments = *ignoreCommentsPtr
	includeComposer = *includeComposerPtr
	source = *sourcePtr // directory/file to start reading MP3 files

  // check for usage, version, or no given source
	if *usagePtr || *versionPtr || source == "" {
		usage()
	}

  // ----- Gather all files
	err = filepath.Walk(source, func(pathedFilename string, info os.FileInfo, err error) error {
		files = append(files, pathedFilename)
		return nil
	})
	if err != nil {
		panic(err)
	}

  // ----- Parse files
	x = 0
	for _, file = range files {
		if strings.Contains(file, ".mp3") {
			filPtr, err = os.Open(file)
			if err != nil {
				fmt.Printf("error loading file (%v): %v", file, err)
				return
			}
			defer filPtr.Close()
			songMetadata, err = tag.ReadFrom(filPtr) // read metadata from file
			x++
			filPtr.Close()

			if err != nil {
				fmt.Printf("error reading file (%v): %v\n", file, err)
				return
			}
      errCode = ""

      // Gather song data
			song.format = string(songMetadata.Format())
			song.idtype = string(songMetadata.FileType())
			song.genre = songMetadata.Genre()
			song.year = songMetadata.Year()
			song.album = songMetadata.Album()
			song.disc, song.discCount = songMetadata.Disc()
			songTrack, song.trackCount = songMetadata.Track()

			if includeComposer {
				song.composer = songMetadata.Composer()
			} else {
				song.composer = ""
			}
			if !ignoreComments {
				song.comment = songMetadata.Comment()
			} else {
				song.comment = ""
			}
			if songMetadata.Picture() != nil {
				pic = songMetadata.Picture()
				song.picture = pic.Ext + "_" + pic.MIMEType + "_" + pic.Type + "_" + pic.Description + "_" + strconv.Itoa(len(pic.Data))
				song.picSize = len(pic.Data)
			} else {
				song.picture = "nil"
				song.picSize = 0
			}

			if songTrack == 1 { // if 1st song of album, use this data as control data for rest of album
				x = 0
				songCtrlVar.format = song.format  // copy the control data...
				songCtrlVar.idtype = song.idtype
				songCtrlVar.album = song.album
				songCtrlVar.genre = song.genre
				songCtrlVar.year = song.year
				songCtrlVar.disc = song.disc
				songCtrlVar.discCount = song.discCount
				songCtrlVar.trackCount = song.trackCount
				songCtrlVar.comment = song.comment
				songCtrlVar.picture = song.picture
				songCtrlVar.picSize = song.picSize
        songMetadataCtrlVarPrinted = false
//        songMetadataCtrlVar = songMetadata
      }

			if song != songCtrlVar || all || (song.picSize < minimumPicSize) {
				if !titlePrinted {
					printTitle()
					titlePrinted = true
				}
        if !songMetadataCtrlVarPrinted {
//				  printMetadata(songMetadataCtrlVar, errCode, x)
          songMetadataCtrlVarPrinted=true
        }

				if song.format != songCtrlVar.format {
          errCode = errCode + "f"
				}
				if song.idtype != songCtrlVar.idtype {
          errCode = errCode + "i"
				}
				if song.genre != songCtrlVar.genre {
          errCode = errCode + "g"
				}
				if song.year != songCtrlVar.year {
          errCode = errCode + "y"
				}
				if song.album != songCtrlVar.album {
          errCode = errCode + "a"
				}
				if song.disc != songCtrlVar.disc {
          errCode = errCode + "d"
				}
				if song.discCount != songCtrlVar.discCount {
          errCode = errCode + "D"
				}
				if song.trackCount != songCtrlVar.trackCount {
          errCode = errCode + "T"
				}
				if !ignoreComments {
					if song.composer != songCtrlVar.composer {
            errCode = errCode + "c"
					}
				}
				if !ignoreComments {
					if song.comment != songCtrlVar.comment {
            errCode = errCode + "C"
					}
				}
				if song.picSize != songCtrlVar.picSize {
          errCode = errCode + "p"
				}
				if song.picSize < minimumPicSize {
          errCode = errCode + "P"
				}
				printMetadata(songMetadata, errCode, x)
			}
		}
	}
}

func printMetadata(m tag.Metadata, errorKey2 string, x int) {
	var disc int
	var discCount int
	var track int
	var trackCount int
	var pic *tag.Picture
	var picInfo string
  var picSize int

	track, trackCount = m.Track()
	disc, discCount = m.Disc()

	if m.Picture() != nil {
    pic = m.Picture()
//		pic = songMetadata.Picture()
//		song.picture = pic.Ext + "_" + pic.MIMEType + "_" + pic.Type + "_" + pic.Description + "_" + strconv.Itoa(len(pic.Data))
    picInfo = pic.Ext + "_" + pic.MIMEType + "_" + pic.Type + "_" + pic.Description + "_" + strconv.Itoa(len(pic.Data))
		picSize = len(pic.Data)
	} else {
		picInfo = "nil"
		picSize = 0
	}

//	pic = m.Picture()
//	picInfo = pic.Ext + "_" + pic.MIMEType + "_" + pic.Type + "_" + pic.Description + "_" + strconv.Itoa(len(pic.Data))

  fmt.Printf("|%16s|", errorKey2)
	fmt.Printf("%4d|", x)
	fmt.Printf("%-8.8s|", m.Format())
	fmt.Printf("%-4.4s|", m.FileType())
	fmt.Printf("%-8.8s|", m.Genre())
	fmt.Printf("%4d|", m.Year())
	fmt.Printf("%-16.16s|", m.Album())
	fmt.Printf("%2d|", disc)
	fmt.Printf("%2d|", discCount)
	fmt.Printf("%2d|", track)
	fmt.Printf("%3d|", trackCount)
	fmt.Printf("%-12.12s|", m.AlbumArtist())
	fmt.Printf("%-12.12s|", m.Artist())
	fmt.Printf("%-16.16s|", m.Title())
	fmt.Printf("%-8.8s|", m.Composer())
	fmt.Printf("%-12.12s|", m.Comment())
	fmt.Printf("%9d|", picSize)
	fmt.Printf("%-16.16s|", picInfo)
	fmt.Printf("\n")
}

func printTitle() {
	fmt.Printf("|%16s| ", "")
	fmt.Printf("%4s|", "")
	fmt.Printf("%-8.8s|", "f")
	fmt.Printf("%-4.4s|", "i")
	fmt.Printf("%-8.8s| ", "g")
	fmt.Printf("%4s|", "y")
	fmt.Printf("%-16.16s|", "b")
	fmt.Printf("%1s|", "d")
	fmt.Printf("%1s|", "D")
	fmt.Printf("%2s|", "t")
	fmt.Printf("%3s|", "T")
	fmt.Printf("%-12.12s| ", "A")
	fmt.Printf("%-12.12s| ", "a")
	fmt.Printf("%-16.16s| ", "s")
	fmt.Printf("%-8.8s| ", "c")
	fmt.Printf("%-12.12s|", "C")
	fmt.Printf("%-9.9s|", "p")
	fmt.Printf("%-16.16s", "P")
	fmt.Printf("\n")

	fmt.Printf("|%16s| ", "err")
	fmt.Printf("%-4.4s|", "x")
	fmt.Printf("%-8.8s|", "Format")
	fmt.Printf("%-4.4s|", "Type")
	fmt.Printf("%-8.8s| ", "Genre")
	fmt.Printf("%4s|", "Year")
	fmt.Printf("%-16.16s| ", "Album")
	fmt.Printf("%1s|", "d")
	fmt.Printf("%1s|", "dc")
	fmt.Printf("%2s|", "t")
	fmt.Printf("%3s|", "tc")
	fmt.Printf("%-12.12s| ", "albmArtist")
	fmt.Printf("%-12.12s| ", "Artist")
	fmt.Printf("%-16.16s| ", "Title")
	fmt.Printf("%-8.8s| ", "Composer")
	fmt.Printf("%-12.12s|", "Comment")
	fmt.Printf("%-9.9s|", "PicSize")
	fmt.Printf("%-16.16s", "Picture")
	fmt.Printf("\n")

  fmt.Println(strings.Repeat("-", 128))

}
