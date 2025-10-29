package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// AlbumSummary holds the summary for each album in the JSON.
type AlbumSummary struct {
	AlbumName string     `json:"album_name"`
	Status    string     `json:"status"` // "ok", "issues", or "warning"
	Summary   struct { // nested anonymous struct for small footprint
		SongsTotal int `json:"songs_total"`
		Issues     int `json:"issues"`
	} `json:"summary"`
	Songs []songInfo `json:"songs"`
}

// FinalReport is the top-level JSON structure.
type FinalReport struct {
	Albums        []AlbumSummary `json:"albums"`
	SummaryTotals struct {
		AlbumsOK         int `json:"albums_ok"`
		AlbumsWithIssues int `json:"albums_with_issues"`
		AlbumsWithWarn   int `json:"albums_with_warnings"`
		SongsTotal       int `json:"songs_total"`
	} `json:"summary_totals"`
	GeneratedAt string `json:"generated_at"`
}

// ANSI color codes (no external dependencies)
const (
	ansiReset  = "\u001b[0m"
	ansiRed    = "\u001b[31m"
	ansiGreen  = "\u001b[32m"
	ansiYellow = "\u001b[33m"
	// bold not required, but available:
	ansiBold = "\u001b[1m"
)

// compareSongs is the same comparison logic used to identify problems between two tracks.
func compareSongs(a, b songInfo, cfg config) string {
	errKey := ""
	if a.Format != b.Format {
		errKey += "f"
	}
	if a.IDType != b.IDType {
		errKey += "i"
	}
	if a.Genre != b.Genre {
		errKey += "g"
	}
	if a.Year != b.Year || b.Year == 0 {
		errKey += "y"
	}
	if a.Album != b.Album {
		errKey += "b"
	}
	if a.Disc != b.Disc {
		errKey += "d"
	}
	if a.DiscCount != b.DiscCount || (cfg.DiscCountZero && b.DiscCount == 0) {
		errKey += "D"
	}
	if a.TrackCount != b.TrackCount || (cfg.TrackCountZero && b.TrackCount == 0) {
		errKey += "T"
	}
	if a.AlbumArtist != b.AlbumArtist {
		errKey += "A"
	}
	if a.Artist != b.Artist {
		errKey += "a"
	}
	if cfg.Composer && a.Composer != b.Composer {
		errKey += "C"
	}
	if cfg.Comment && a.Comment != b.Comment {
		errKey += "c"
	}
	if cfg.Picture && a.PicSize != b.PicSize {
		errKey += "p"
	}
	if cfg.Picture && b.PicSize < cfg.MinPicSize {
		errKey += "P"
	}
	return errKey
}

// BuildAlbumReports groups songs by album, computes per-album status, and returns FinalReport + flattened results for CSV.
func BuildAlbumReports(songs []songInfo, cfg config) (FinalReport, []songInfo) {
	report := FinalReport{}
	albumMap := map[string][]songInfo{}
	totalSongs := 0

	// Group songs by album (normalized album name)
	for _, s := range songs {
		name := strings.TrimSpace(s.Album)
		if name == "" {
			// fallback: use directory name if album tag missing
			dir := filepath.Dir(s.Path)
			name = filepath.Base(dir)
			if name == "." || name == "" {
				name = "<unknown>"
			}
		}
		albumMap[name] = append(albumMap[name], s)
		totalSongs++
	}

	// Build album summaries
	albumsOK := 0
	albumsIssues := 0
	albumsWarn := 0


for albumName, list := range albumMap {
    var alb AlbumSummary
    alb.AlbumName = albumName
    alb.Songs = make([]songInfo, 0, len(list))
    alb.Summary.SongsTotal = len(list)
    alb.Summary.Issues = 0

    baseline := list[0]
    for _, s := range list {
        errKey := compareSongs(baseline, s, cfg)
        s.ErrorKey = errKey
        alb.Songs = append(alb.Songs, s)
        if errKey != "" {
            alb.Summary.Issues++
        }
    }

    // Determine album status
    if alb.Summary.Issues > 0 {
        alb.Status = "issues"
        albumsIssues++
    } else {
        warns := 0
        for _, s := range alb.Songs {
            if cfg.Picture && s.PicSize == 0 {
                warns++
            }
            if cfg.TrackCountZero && s.TrackCount == 0 {
                warns++
            }
            if s.Year == 0 {
                warns++
            }
        }
        if warns > 0 {
            alb.Status = "warning"
            albumsWarn++
        } else {
            alb.Status = "ok"
            albumsOK++
        }
    }

    // Append if album has any relevance OR -all flag is set
    if cfg.AllAll || cfg.All || alb.Status != "ok" {
        report.Albums = append(report.Albums, alb)
    }
}




	// finalize totals
	report.SummaryTotals.AlbumsOK = albumsOK
	report.SummaryTotals.AlbumsWithIssues = albumsIssues
	report.SummaryTotals.AlbumsWithWarn = albumsWarn
	report.SummaryTotals.SongsTotal = totalSongs
	report.GeneratedAt = time.Now().Format(time.RFC3339)

	// Collect flattened result for CSV (we'll include ErrorKey per row)
	flat := make([]songInfo, 0, totalSongs)
	for _, alb := range report.Albums {
		for _, s := range alb.Songs {
			flat = append(flat, s)
		}
	}

	return report, flat
}


func PrintANSIAlbumSummary(r FinalReport, cfg config) {
	hline := strings.Repeat("─", 55)
	fmt.Println()
	fmt.Println(hline)

	// 🔹 Error Code Legend
	fmt.Printf("%sLegend:%s ", ansiBold, ansiReset)
	fmt.Printf("%sf%s=format ", ansiRed, ansiReset)
	fmt.Printf("%si%s=id ", ansiRed, ansiReset)
	fmt.Printf("%sg%s=genre ", ansiRed, ansiReset)
	fmt.Printf("%sy%s=year ", ansiRed, ansiReset)
	fmt.Printf("%sb%s=album ", ansiRed, ansiReset)
	fmt.Printf("%sd%s=disc ", ansiRed, ansiReset)
	fmt.Printf("%sD%s=disc# ", ansiRed, ansiReset)
	fmt.Printf("%sT%s=track# ", ansiRed, ansiReset)
	fmt.Printf("%sA%s=albumArtist ", ansiRed, ansiReset)
	fmt.Printf("%sa%s=artist ", ansiRed, ansiReset)
	fmt.Printf("%sC%s=composer ", ansiRed, ansiReset)
	fmt.Printf("%sc%s=comment ", ansiRed, ansiReset)
	fmt.Printf("%sp%s=picSize ", ansiRed, ansiReset)
	fmt.Printf("%sP%s=minPic\n", ansiRed, ansiReset)


	fmt.Println(" Album Summary")
	fmt.Println(hline)

	for _, a := range r.Albums {
		switch a.Status {
		case "ok":
			fmt.Printf(" %s✅%s %-30s | %3d songs | OK\n",
				ansiGreen, ansiReset, padRight(a.AlbumName, 30), a.Summary.SongsTotal)

//jwi add allall
		    if cfg.AllAll {
		        for _, s := range a.Songs {
		            fmt.Printf("   ↳ %-30s  artist: %-20s albumArtist: %-20s\n",
		                padRight(s.Title, 30), s.Artist, s.AlbumArtist)
		        }
			}



		case "issues":
			fmt.Printf(" %s⛔%s %-30s | %3d songs | ISSUES\n",
				ansiRed, ansiReset, padRight(a.AlbumName, 30), a.Summary.SongsTotal)

			// show the individual problem tracks
			for _, s := range a.Songs {
				if s.ErrorKey != "" {
					fmt.Printf("   ↳ %-30s  issues: %s%s%s\n",
						padRight(s.Title, 30), ansiRed, s.ErrorKey, ansiReset)
				}
			}

		case "warning":
			fmt.Printf(" %s⚠️ %s %-30s | %3d songs | WARNINGS\n",
				ansiYellow, ansiReset, padRight(a.AlbumName, 30), a.Summary.SongsTotal)

			for _, s := range a.Songs {
				if s.ErrorKey == "" && (s.PicSize == 0 || s.Year == 0 || s.TrackCount == 0) {
					fmt.Printf("   ↳ %-30s  warning: ", padRight(s.Title, 30))
					if s.PicSize == 0 {
						fmt.Printf("missing cover ")
					}
					if s.Year == 0 {
						fmt.Printf("no year ")
					}
					if s.TrackCount == 0 {
						fmt.Printf("no track count ")
					}
					fmt.Println()
				}
			}

		default:
			fmt.Printf(" %-30s | %3d songs | %s\n",
				padRight(a.AlbumName, 30), a.Summary.SongsTotal, a.Status)
		}
	}

	fmt.Println(hline)

//jwi
if r.SummaryTotals.AlbumsWithIssues == 0 && r.SummaryTotals.AlbumsWithWarn == 0 {
    fmt.Printf("%s🎉 All albums are consistent — no issues found!%s\n", ansiGreen, ansiReset)
} else {
//jwi eo new
	fmt.Printf(" %s✅ %d OK albums%s, %s❌ %d with issues%s, %s⚠️ %d warnings%s\n",
		ansiGreen, r.SummaryTotals.AlbumsOK, ansiReset,
		ansiRed, r.SummaryTotals.AlbumsWithIssues, ansiReset,
		ansiYellow, r.SummaryTotals.AlbumsWithWarn, ansiReset)
//jwi one more line new stuff
}
	fmt.Printf(" Total songs: %d\n", r.SummaryTotals.SongsTotal)
	fmt.Println(hline)
}




// PrintANSIAlbumSummary prints the album summary table to terminal using ANSI colors.
func PrintANSIAlbumSummaryOLD(r FinalReport) {
	// header
	hline := strings.Repeat("─", 55)
	fmt.Println()
	fmt.Println(hline)
	fmt.Println(" Album Summary")
	fmt.Println(hline)

	// each album
	for _, a := range r.Albums {
		switch a.Status {
		case "ok":
			fmt.Printf(" %s\u2B55%s %-30s | %3d songs | OK\n", ansiGreen, ansiReset, padRight(a.AlbumName, 30), a.Summary.SongsTotal)
		case "issues":
			fmt.Printf(" %s\u26D4%s %-30s | %3d songs | %s\n", ansiRed, ansiReset, padRight(a.AlbumName, 30), a.Summary.SongsTotal, "ISSUES")
		case "warning":
			fmt.Printf(" %s\u26A0%s %-30s | %3d songs | %s\n", ansiYellow, ansiReset, padRight(a.AlbumName, 30), a.Summary.SongsTotal, "WARNINGS")
		default:
			fmt.Printf(" %-30s | %3d songs | %s\n", padRight(a.AlbumName, 30), a.Summary.SongsTotal, a.Status)
		}
	}

	fmt.Println(hline)
	// totals line with colors
	fmt.Printf(" %s✅ %d OK albums%s, %s❌ %d with issues%s, %s⚠️ %d warnings%s\n",
		ansiGreen, r.SummaryTotals.AlbumsOK, ansiReset,
		ansiRed, r.SummaryTotals.AlbumsWithIssues, ansiReset,
		ansiYellow, r.SummaryTotals.AlbumsWithWarn, ansiReset)
	fmt.Printf(" Total songs: %d\n", r.SummaryTotals.SongsTotal)
	fmt.Println(hline)
}

// padRight ensures fixed-width columns for nice alignment
func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// WriteJSONReport writes the structured album-based JSON report into the current working directory as tag_report.json
func WriteJSONReport(report FinalReport, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("error creating json report: %v", err)
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		log.Printf("error encoding json report: %v", err)
		return
	}
	fmt.Printf("✅ JSON report written to %s\n", filename)
}

// WriteCSVReport writes a flattened CSV report into the current working directory as tag_report.csv
func WriteCSVReport(flat []songInfo, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("error creating csv report: %v", err)
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// header
	header := []string{"path", "album", "artist", "title", "year", "track", "trackCount", "disc", "discCount", "albumArtist", "composer", "comment", "picSize", "errorKey"}
	if err := w.Write(header); err != nil {
		log.Printf("error writing csv header: %v", err)
		return
	}

	for _, s := range flat {
		row := []string{
			s.Path, s.Album, s.Artist, s.Title,
			strconv.Itoa(s.Year), strconv.Itoa(s.Track), strconv.Itoa(s.TrackCount),
			strconv.Itoa(s.Disc), strconv.Itoa(s.DiscCount), s.AlbumArtist,
			s.Composer, s.Comment, strconv.Itoa(s.PicSize), s.ErrorKey,
		}
		if err := w.Write(row); err != nil {
			log.Printf("error writing csv row: %v", err)
		}
	}
	fmt.Printf("✅ CSV report written to %s\n", filename)
}


func WriteReports(songs []songInfo, cfg config) {
    if len(songs) == 0 {
        fmt.Println("No songs to report.")
        return
    }

    // Build summarized + flat reports
    final, flat := BuildAlbumReports(songs, cfg)

    // Determine home directory target
    homeDir, err := os.UserHomeDir()
    if err != nil {
        fmt.Println("⚠️  Could not resolve home directory, using current folder")
        homeDir = "."
    }

    jsonPath := filepath.Join(homeDir, "tag_report.json")
    csvPath := filepath.Join(homeDir, "tag_report.csv")

    fmt.Printf("\n📝 Writing reports to %s ...\n", homeDir)

    // Use your existing writers
    WriteJSONReport(final, jsonPath)
    WriteCSVReport(flat, csvPath)

    // Print album summary in ANSI colors
    PrintANSIAlbumSummary(final, cfg)

    fmt.Println("──────────────────────────────")
    fmt.Printf(" Reports written to:\n   %s\n   %s\n", jsonPath, csvPath)
    fmt.Println("──────────────────────────────")
}

