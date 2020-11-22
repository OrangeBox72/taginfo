/*
file:	 mp3taginfo
author: johnny
descr:	The tag tool reads metadata from media files (as supported by the tag library).
usage:	mp3taginfo -src=<directory/file>
*/
package main

import (
	/*	"encoding/json"	*/
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/OrangeBox72/mp3tagInfo/version"
	"github.com/dhowden/tag"
)
// ===== STRUCTS =======================================================
type songInfo struct {
	format			string
	idtype			string
	genre				string
	year				int
	album				string
	disc				int
	discCount		int
	track				int
	trackCount	int
	albumArtist	string
	artist			string
	title				string
	composer		string
	comment			string
	picture			string
	picSize			int
}

// ===== FUNCTIONS =====================================================
func usage() {
	fmt.Printf("%v - Version: %v\n", os.Args[0], version.BuildVersion)
	fmt.Fprintf(os.Stderr, "\n")
	flag.PrintDefaults()
	os.Exit(0)
}

// --------------------
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// --------------------
func getSongInfo(smd tag.Metadata) songInfo {
	var s songInfo
	var pic *tag.Picture

	s.format = string(smd.Format())
	s.idtype = string(smd.FileType())
	s.genre = smd.Genre()
	s.year = smd.Year()
	s.album = smd.Album()
	s.disc, s.discCount = smd.Disc()
	s.track, s.trackCount = smd.Track()
	s.albumArtist = smd.AlbumArtist()
	s.artist = smd.Artist()
	s.title = smd.Title()
	s.composer = smd.Composer()
	s.comment = smd.Comment()
	if smd.Picture() != nil {
		pic = smd.Picture()
		s.picture = pic.Ext + "_" + pic.MIMEType + "_" + pic.Type + "_" + pic.Description + "_" + strconv.Itoa(len(pic.Data))
		s.picSize = len(pic.Data)
	} else {
		s.picture = "nil"
		s.picSize = 0
	}
	return s
}

// --------------------
func maskedSongInfo(s songInfo) songInfo {				// this masks out unique info
	s.track = 0
	s.albumArtist = ""
	s.artist = ""
	s.title = ""
	if !composer {
		s.composer = ""
	}
	if !comment {
		s.comment = ""
	}
	return s
}

// --------------------
func printSong(s songInfo, errorKey string, x int) {
	fmt.Printf("|%16s|", errorKey)
	fmt.Printf("%4d|", x)
	fmt.Printf("%-8.8s|", s.format)
	fmt.Printf("%-4.4s|", s.idtype)
	fmt.Printf("%-8.8s|", s.genre)
	fmt.Printf("%4d|", s.year)
	fmt.Printf("%-16.16s|", s.album)
	fmt.Printf("%3d|", s.disc)
	fmt.Printf("%3d|", s.discCount)
	fmt.Printf("%3d|", s.track)
	fmt.Printf("%3d|", s.trackCount)
	fmt.Printf("%-12.12s|", s.albumArtist)
	fmt.Printf("%-12.12s|", s.artist)
	fmt.Printf("%-16.16s|", s.title)
	fmt.Printf("%-8.8s|", s.composer)
	fmt.Printf("%-12.12s|", s.comment)
	fmt.Printf("%9d|", s.picSize)
	fmt.Printf("%-16.16s|", s.picture)
	fmt.Printf("\n")
}

// --------------------
func printTitle() {
	var l = 176

	fmt.Println(strings.Repeat("-", l))
	fmt.Printf("|%16s|", "err")
	fmt.Printf("%-4.4s|", "x")
	fmt.Printf("%-8.8s|", "(f)ormat")
	fmt.Printf("%-4.4s|", "(i)d")
	fmt.Printf("%-8.8s|", "(g)enre")
	fmt.Printf("%4s|", "(y)r")
	fmt.Printf("%-16.16s|", "al(b)um")
	fmt.Printf("%3s|", "(d)")
	fmt.Printf("%3s|", "(D)")
	fmt.Printf("%3s|", "(t)")
	fmt.Printf("%3s|", "(T)")
	fmt.Printf("%-12.12s|", "albm(A)rtst")
	fmt.Printf("%-12.12s|", "(a)rtist")
	fmt.Printf("%-16.16s|", "(t)itle")
	fmt.Printf("%-8.8s|", "(c)ompsr")
	fmt.Printf("%-12.12s|", "(C)omment")
	fmt.Printf("%-9.9s|", "(P)icSize")
	fmt.Printf("%-16.16s", "(p)icture")
	fmt.Printf("\n")

	fmt.Println(strings.Repeat("-", l))
}



// ===== GLOBAL VARS ===================================================
var comment bool
var composer bool
var disccount bool
var picture bool
var trackcount bool





// ===== MAIN ==========================================================
func main() {
	// Vars ------------------------
	var filPtr *os.File
	var songMetadata tag.Metadata
	var comparedSong songInfo
	var song songInfo
//	var cleanedSong songInfo
//	var cleanedComparedSong songInfo

	var err error
	var files []string
	var file string

	// commandline vars
	var allPtr *bool
	var all bool
	var commentPtr *bool
	var composerPtr *bool
	var disccountPtr *bool
	var minimumPicSizePtr *int
	var picturePtr *bool
	var trackcountPtr *bool
	var sourcePtr *string
	var source string
	var usagePtr *bool
	var versionPtr *bool

	var albumName string
	var errCode string
//	var songTrack int
	var x int
	var titlePrinted bool
	var song1Printed bool
	var minimumPicSize int

	titlePrinted = false
	song1Printed = false
//	minimumPicSize = 4500
	// ----- Passed Args ----------
	allPtr = flag.Bool("all", false, "Prints info on ALL files regardless of Tag irregularities")
	commentPtr = flag.Bool("comment", false, "compares *comment* fields")
	composerPtr = flag.Bool("composer", false, "compares *composer* fields")
	disccountPtr = flag.Bool("disccountzero", false, "checks for EMPTY *disccount*")
	minimumPicSizePtr = flag.Int("minimumpicsize", 8000, "compares *picSize* fields against this minimum value" )
	picturePtr = flag.Bool("picture", true, "compares *picture* fields")
	sourcePtr = flag.String("src", "", "<REQUIRED> Source of MP3's to parse for Tags")
	trackcountPtr = flag.Bool("trackcountzero", false, "checks for EMPTY *trackcount*")
	usagePtr = flag.Bool("usage", false, "This message")
	versionPtr = flag.Bool("version", false, "Version info")
	flag.Parse()

	all = *allPtr
	comment = *commentPtr
	composer = *composerPtr
	disccount = *disccountPtr
	minimumPicSize = *minimumPicSizePtr
	picture = *picturePtr
	trackcount = *trackcountPtr
	source = *sourcePtr

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
	albumName = ""
	errCode = ""
	for _, file = range files {
		if strings.Contains(file, ".mp3") {
			filPtr, err = os.Open(file)
			if err != nil {
				fmt.Printf("error loading file (%v): %v", file, err)
				return
			}
			defer filPtr.Close()
			songMetadata, err = tag.ReadFrom(filPtr) // read metadata from file
			filPtr.Close()

			if err != nil {
				fmt.Printf("error reading file (%v): %v\n", file, err)
				return
			}
//			errCode = ""

			// Gather song data
			x++
			song = getSongInfo(songMetadata)
			if albumName != song.album {				// ie processing a new ablum
				//  set your control var data for album
				comparedSong = song
//RIGHT HERE  <--- I HAVE NO CLUE NOW.. WHAT I WAS DOING HERE
				albumName = comparedSong.album
				errCode = ""
			} else {    // if they are the same
				// compare each file/record against the control var data
				if (maskedSongInfo(comparedSong) != maskedSongInfo(song)) || all || (picture && (song.picSize < minimumPicSize)) || trackcount || disccount {
					errCode = ""
					if song.format != comparedSong.format {
						errCode = errCode + "f"
					}
					if song.idtype != comparedSong.idtype {
						errCode = errCode + "i"
					}
					if song.genre != comparedSong.genre {
						errCode = errCode + "g"
					}
					if song.year != comparedSong.year {
						errCode = errCode + "y"
					}
					if song.album != comparedSong.album {
						errCode = errCode + "a"
					}
					if song.disc != comparedSong.disc {
						errCode = errCode + "d"
					}
					if (song.discCount != comparedSong.discCount) || (disccount && song.discCount == 0) {
						if !strings.Contains(errCode, "D") {
							errCode = errCode + "D"
						}
					}
					if (song.trackCount != comparedSong.trackCount) || (trackcount && song.trackCount == 0) {
						if !strings.Contains(errCode, "T") {
							errCode = errCode + "T"
						}
					}
					if song.composer != comparedSong.composer && composer {
						errCode = errCode + "c"
					}
					if song.comment != comparedSong.comment && comment {
						errCode = errCode + "C"
					}
					if song.picSize != comparedSong.picSize && picture {
						errCode = errCode + "p"
					}
					if song.picSize < minimumPicSize && picture {
						errCode = errCode + "P"
					}
				} //eoif triggered Masked songs dont equal
				if errCode > "" || all {
					if !titlePrinted {
						printTitle()
						titlePrinted = true
					}
					if !song1Printed {
						printSong(comparedSong, "", x-1)
						song1Printed = true
					}
					printSong(song, errCode, x)
				}
//				} //eoif triggered Masked songs dont equal
			} //eoElse songs are same
		} //eoif mp3 filetype
	} // eoif range of files traverse
} // eoif Main
