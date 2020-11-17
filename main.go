/*
file:   mp3taginfo
author: johnny
descr:  The tag tool reads metadata from media files (as supported by the tag library).
usage:  mp3taginfo <directory/file>
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

	"github.com/dhowden/tag"
)

// ===== FUNCTIONS =====================================================
func usage() {
	fmt.Fprintf(os.Stderr, "mp3taginfo\n")
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
	var directory string
	var filPtr *os.File
	var songMetadata tag.Metadata
	var songMetadata1 tag.Metadata
	var song1 songCommonInfo
	var songX songCommonInfo
	//	var pic1 *tag.Picture
	var picX *tag.Picture
	var err error
	var files []string
	var file string
	var usagePtr *bool
	var directoryPtr *string
	var printItPtr *bool
	var ignoreCommentsPtr *bool
	var includeComposerPtr *bool
	var printIt bool
	var ignoreComments bool
	var includeComposer bool
	var errMrk int
	var trackX int
	var x int

	// Passed Args
	usagePtr = flag.Bool("usage", false, "This message")
	directoryPtr = flag.String("dir", "./", "Directory of MP3's to parse for Tags")
	printItPtr = flag.Bool("print", false, "Prints details about artist/album/song")
	ignoreCommentsPtr = flag.Bool("ignore-comments", false, "ignores comment fields in comparisons")
	includeComposerPtr = flag.Bool("include-composer", false, "includes composer fields in comparisons")
	flag.Parse()

	if *usagePtr { // if '-usage' given.. then show Usage and exit
		usage()
	}

	directory = *directoryPtr // directory to start reading MP3 files
	printIt = *printItPtr
	ignoreComments = *ignoreCommentsPtr
	includeComposer = *includeComposerPtr

	err = filepath.Walk(directory, func(pathedFilename string, info os.FileInfo, err error) error {
		files = append(files, pathedFilename)
		return nil
	})
	if err != nil {
		panic(err)
	}

	errMrk = 0
	x = 0
	printTitle()
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
			if printIt {
				printMetadata(songMetadata, -1, x)
			} else {
				errMrk = 0
				songX.format = string(songMetadata.Format())
				songX.idtype = string(songMetadata.FileType())
				songX.genre = songMetadata.Genre()
				songX.year = songMetadata.Year()
				songX.album = songMetadata.Album()
				songX.disc, songX.discCount = songMetadata.Disc()
				trackX, songX.trackCount = songMetadata.Track()
				if includeComposer {
					fmt.Println("FALSE")
					songX.composer = songMetadata.Composer()
				} else {
					songX.composer = ""
				}
				if !ignoreComments {
					songX.comment = songMetadata.Comment()
				} else {
					songX.comment = ""
				}
				picX = songMetadata.Picture()
				songX.picture = picX.Ext + "_" + picX.MIMEType + "_" + picX.Type + "_" + picX.Description + "_" + strconv.Itoa(len(picX.Data))
				songX.picSize = len(picX.Data)

				if trackX == 1 { // if 1st song of album, use this data as control data for rest of album
					x = 0
					songMetadata1 = songMetadata // copy this just in case we do an err report later
					song1.format = songX.format  // copy the control data...
					song1.idtype = songX.idtype
					song1.album = songX.album
					song1.genre = songX.genre
					song1.year = songX.year
					song1.disc = songX.disc
					song1.discCount = songX.discCount
					song1.trackCount = songX.trackCount
					song1.comment = songX.comment
					song1.picture = songX.picture
					song1.picSize = songX.picSize

				} else {
					if songX != song1 {
						if trackX == 2 {
							printMetadata(songMetadata1, 0, x)
						}
						if songX.format != song1.format {
							errMrk = errMrk + 1
						}
						if songX.idtype != song1.idtype {
							errMrk = errMrk + 2
						}
						if songX.genre != song1.genre {
							errMrk = errMrk + 4
						}
						if songX.year != song1.year {
							errMrk = errMrk + 8
						}
						if songX.album != song1.album {
							errMrk = errMrk + 16
						}
						if songX.disc != song1.disc {
							errMrk = errMrk + 32
						}
						if songX.discCount != song1.discCount {
							errMrk = errMrk + 64
						}
						if songX.trackCount != song1.trackCount {
							errMrk = errMrk + 128
						}
						if !ignoreComments {
							if songX.composer != song1.composer {
								errMrk = errMrk + 256
							}
						}
						if !ignoreComments {
							if songX.comment != song1.comment {
								errMrk = errMrk + 512
							}
						}
						if songX.picSize != song1.picSize {
							errMrk = errMrk + 1024
						}
						printMetadata(songMetadata, errMrk, x)
					}
				}
				x++
			}
		}
	}
}

func printMetadata(m tag.Metadata, errorKey int, x int) {
	var disc int
	var discCount int
	var track int
	var trackCount int
	var pic *tag.Picture
	var picInfo string

	track, trackCount = m.Track()
	disc, discCount = m.Disc()
	pic = m.Picture()
	picInfo = pic.Ext + "_" + pic.MIMEType + "_" + pic.Type + "_" + pic.Description + "_" + strconv.Itoa(len(pic.Data))

	fmt.Printf("%4d| ", errorKey)
	fmt.Printf("%4d|", x)
	fmt.Printf("%-8.8s|", m.Format())
	fmt.Printf("%-4.4s|", m.FileType())
	fmt.Printf("%-8.8s| ", m.Genre())
	fmt.Printf("%4d|", m.Year())
	fmt.Printf("%-16.16s|", m.Album())
	fmt.Printf("%2d|", disc)
	fmt.Printf("%2d|", discCount)
	fmt.Printf("%2d|", track)
	fmt.Printf("%3d|", trackCount)
	fmt.Printf("%-12.12s| ", m.AlbumArtist())
	fmt.Printf("%-12.12s| ", m.Artist())
	fmt.Printf("%-16.16s| ", m.Title())
	fmt.Printf("%-8.8s| ", m.Composer())
	fmt.Printf("%-12.12s|", m.Comment())
	fmt.Printf("%8d|", len(pic.Data))
	fmt.Printf("%-16.16s|", picInfo)
	fmt.Printf("\n")
}

func printTitle() {
	fmt.Printf("%4s| ", "")
	fmt.Printf("%4s|", "")
	fmt.Printf("%-8.8s|", "1")
	fmt.Printf("%-4.4s|", "2")
	fmt.Printf("%-8.8s| ", "4")
	fmt.Printf("%4s|", "8")
	fmt.Printf("%-16.16s|", "16")
	fmt.Printf("%1s|", "32")
	fmt.Printf("%1s|", "64")
	fmt.Printf("%2s|", "")
	fmt.Printf("%3s|", "128")
	fmt.Printf("%-12.12s| ", "")
	fmt.Printf("%-12.12s| ", "")
	fmt.Printf("%-16.16s| ", "")
	fmt.Printf("%-8.8s| ", "256")
	fmt.Printf("%-12.12s|", "512")
	fmt.Printf("%-8.8s|", "1024")
	fmt.Printf("%-16.16s", "")
	fmt.Printf("\n")

	fmt.Printf("%4s| ", "err")
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
	fmt.Printf("%-8.8s|", "PicSize")
	fmt.Printf("%-16.16s", "Picture")
	fmt.Printf("\n")
}
