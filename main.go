/*

example usage -  taginfo -src=./Music -json -csv -workers=12

*/
package main

	// "path/filepath"
import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Printf("%s - Version: %s\n", os.Args[0], "2.1-modular")
	fmt.Println("Scans audio files (mp3/m4a/alac), compares tags across an album,")
	fmt.Println("and produces JSON/CSV reports in the current working directory.")
	fmt.Println()
	fmt.Println("Usage example:")
	fmt.Printf("  %s -src=./Music -workers=8 -json -csv\n\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	// flags
	var cfg config
	flag.BoolVar(&cfg.All, "all", false, "Show all albums (include those without irregularities)")
	flag.BoolVar(&cfg.AllAll, "allall", false, "Show all files (even those without irregularities)")
	flag.BoolVar(&cfg.Comment, "comment", false, "Consider comment fields when comparing")
	flag.BoolVar(&cfg.Composer, "composer", false, "Consider composer fields when comparing")
	flag.BoolVar(&cfg.DiscCountZero, "disccountzero", true, "Flag empty disc counts as issues")
	flag.BoolVar(&cfg.TrackCountZero, "trackcountzero", true, "Flag empty track counts as issues")
	flag.BoolVar(&cfg.Picture, "picture", true, "Compare picture fields")
	flag.IntVar(&cfg.MinPicSize, "minimumpicsize", 8000, "Minimum acceptable picture size")
	flag.StringVar(&cfg.Source, "src", "", "Source directory to scan (required)")
//	quietPtr := flag.Bool("quiet", false, "Suppress per-song progress output")
	jsonFlag := flag.Bool("json", false, "Write JSON report (tag_report.json in PWD)")
	csvFlag := flag.Bool("csv", false, "Write CSV report (tag_report.csv in PWD)")
	flag.IntVar(&cfg.Workers, "workers", 8, "Number of parallel workers (goroutines)")
	help := flag.Bool("help", false, "Show usage")
	flag.Parse()

	//johnny vars
//	var fqfn, app string
//	fqfn=os.Args[0]
//	app=filepath.Base(fqfn)

	if *help || cfg.Source == "" {
		usage()
	}

	cfg.JSON = *jsonFlag
	cfg.CSV = *csvFlag
	
//	quiet := *quietPtr



	//jwi find homedir
//	homeDir, err := os.UserHomeDir()
//	if err != nil {
//    	fmt.Println("⚠️  Could not resolve home directory, using current folder")
//    	homeDir = "."
//	}
//	jsonPath := filepath.Join(homeDir, fmt.Sprintf("%s_report.json", app))
//	csvPath := filepath.Join(homeDir, fmt.Sprintf("%s_report.csv", app))



	// ensure source exists
	if _, err := os.Stat(cfg.Source); os.IsNotExist(err) {
		fmt.Printf("source path does not exist: %s\n", cfg.Source)
		os.Exit(1)
	}

	// Run scanner
	fmt.Printf("🎵 Scanning: %s  (workers=%d)\n", cfg.Source, cfg.Workers)
	start := Now()

	songs := ScanFiles(cfg.Source, cfg.Workers)

//	fmt.Printf("\n📦 Processed %d files\n", len(songs))

	// After scanning all songs
	fmt.Printf("\n📦 Processed %d files\n", len(songs))

	// Build and write reports (writes JSON + CSV + summary)
	WriteReports(songs, cfg)

	// Done
	fmt.Printf("⏱️  Elapsed: %v\n", Now().Sub(start).Round(1e7))
}
