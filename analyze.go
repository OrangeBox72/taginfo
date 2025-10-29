package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/dhowden/tag"
)

// scanResult represents either a successful song read or an error.
type scanResult struct {
	Song songInfo
	Err  error
}

// walkCollectFiles recursively collects all audio files under src.
// It supports .mp3, .m4a, .alac (case-insensitive).
func walkCollectFiles(src string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// propagate error walking into folder (permission, etc.)
			return err
		}
		if info.IsDir() {
			return nil
		}
		l := strings.ToLower(path)
		if strings.HasSuffix(l, ".mp3") || strings.HasSuffix(l, ".m4a") || strings.HasSuffix(l, ".alac") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// ScanFiles reads tags from all files in src using a worker pool.
// It prints simple per-file progress lines in the form:
//   [  1/120] Scanning: /path/to/file.mp3
//
// Returns slice of songInfo for all successfully read files.
func ScanFiles(src string, workers int) []songInfo {
	files, err := walkCollectFiles(src)
	if err != nil {
		log.Fatalf("error walking directory %s: %v", src, err)
	}

	total := len(files)
	if total == 0 {
		fmt.Printf("No audio files found in %s\n", src)
		return nil
	}

	results := make([]songInfo, 0, total)
	var mu sync.Mutex
	var wg sync.WaitGroup
	pathCh := make(chan string)

	var processed int64 = 0

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range pathCh {
				// Open file and read tag metadata
				f, err := os.Open(p)
				if err != nil {
					atomic.AddInt64(&processed, 1)
					fmt.Printf("[%3d/%d] Scanning: %s  (error opening: %v)\n", int(atomic.LoadInt64(&processed)), total, p, err)
					continue
				}
				meta, err := tag.ReadFrom(f)
				f.Close()
				if err != nil {
					atomic.AddInt64(&processed, 1)
					fmt.Printf("[%3d/%d] Scanning: %s  (error reading tags: %v)\n", int(atomic.LoadInt64(&processed)), total, p, err)
					continue
				}

				s := getSongInfo(p, meta)

				// store
				mu.Lock()
				results = append(results, s)
				mu.Unlock()

				// progress line (simple lines)
				processedCount := int(atomic.AddInt64(&processed, 1))
				fmt.Printf("[%3d/%d] Scanning: %s\n", processedCount, total, p)
			}
		}()
	}

	// feed paths
	go func() {
		for _, p := range files {
			pathCh <- p
		}
		close(pathCh)
	}()
	wg.Wait()
	return results
}

