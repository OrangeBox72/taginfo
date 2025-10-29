package main

import (
	"fmt"
	"github.com/dhowden/tag"
)

// songInfo holds metadata for a single audio file.
type songInfo struct {
	Path        string `json:"path"`
	Format      string `json:"format"`
	IDType      string `json:"idtype"`
	Genre       string `json:"genre"`
	Year        int    `json:"year"`
	Album       string `json:"album"`
	Disc        int    `json:"disc"`
	DiscCount   int    `json:"disc_count"`
	Track       int    `json:"track"`
	TrackCount  int    `json:"track_count"`
	AlbumArtist string `json:"album_artist"`
	Artist      string `json:"artist"`
	Title       string `json:"title"`
	Composer    string `json:"composer"`
	Comment     string `json:"comment"`
	Picture     string `json:"picture"`
	PicSize     int    `json:"pic_size"`
	ErrorKey    string `json:"error_key"`
}

// getSongInfo reads tags from a file and returns songInfo.
// Uses github.com/dhowden/tag.
func getSongInfo(path string, meta tag.Metadata) songInfo {
	var s songInfo
	s.Path = path
	s.Format = string(meta.Format())
	s.IDType = string(meta.FileType())
	s.Genre = meta.Genre()
	s.Year = meta.Year()
	s.Album = meta.Album()
	s.Disc, s.DiscCount = meta.Disc()
	s.Track, s.TrackCount = meta.Track()
	s.AlbumArtist = meta.AlbumArtist()
	s.Artist = meta.Artist()
	s.Title = meta.Title()
	s.Composer = meta.Composer()
	s.Comment = meta.Comment()

	if pic := meta.Picture(); pic != nil {
		s.Picture = fmt.Sprintf("%s_%s_%s_%s_%d", pic.Ext, pic.MIMEType, pic.Type, pic.Description, len(pic.Data))
		s.PicSize = len(pic.Data)
	}
	return s
}

