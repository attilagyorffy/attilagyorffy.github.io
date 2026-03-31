package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
	thumbSize    = 600
	fullSize     = 2000
	thumbQuality = 80
	fullQuality  = 85
)

// optimizePhotos converts raw photo exports (JPEG/PNG) from an album source
// directory into web-optimised WebP files under public/photos/<slug>/web/.
// It shells out to ImageMagick's `magick` command for the actual conversion.
func optimizePhotos(root, albumSlug string) error {
	srcDir := filepath.Join(root, "photos", albumSlug)
	outDir := filepath.Join(root, "public", "photos", albumSlug, "web")
	thumbDir := filepath.Join(outDir, "thumb")
	fullDir := filepath.Join(outDir, "full")

	if err := os.MkdirAll(thumbDir, 0o755); err != nil {
		return fmt.Errorf("creating thumb dir: %w", err)
	}
	if err := os.MkdirAll(fullDir, 0o755); err != nil {
		return fmt.Errorf("creating full dir: %w", err)
	}

	// Collect source images (JPEG/PNG).
	sources, err := collectSourcePhotos(srcDir)
	if err != nil {
		return fmt.Errorf("collecting photos from %s: %w", srcDir, err)
	}
	if len(sources) == 0 {
		fmt.Printf("  \033[33m[photos]\033[0m no source images in %s\n", srcDir)
		return nil
	}

	// Process images concurrently (up to 4 at a time).
	var (
		wg      sync.WaitGroup
		sem     = make(chan struct{}, 4)
		mu      sync.Mutex
		firstErr error
	)

	for i, src := range sources {
		padded := fmt.Sprintf("%03d", i+1)

		wg.Add(1)
		go func(src, padded string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fmt.Printf("  \033[36m[photos]\033[0m [%s] %s\n", padded, filepath.Base(src))

			if err := convertImage(src, filepath.Join(thumbDir, padded+".webp"), thumbSize, thumbQuality); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("thumb %s: %w", padded, err)
				}
				mu.Unlock()
				return
			}

			if err := convertImage(src, filepath.Join(fullDir, padded+".webp"), fullSize, fullQuality); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("full %s: %w", padded, err)
				}
				mu.Unlock()
			}
		}(src, padded)
	}

	wg.Wait()
	if firstErr != nil {
		return firstErr
	}

	fmt.Printf("  \033[32m[photos]\033[0m %s: %d images optimised\n", albumSlug, len(sources))
	return nil
}

// collectSourcePhotos returns sorted paths to JPEG/PNG files in the directory.
func collectSourcePhotos(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var photos []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".jpeg" || ext == ".jpg" || ext == ".png" {
			photos = append(photos, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(photos)
	return photos, nil
}

// convertImage uses ImageMagick to resize and convert an image to WebP.
func convertImage(src, dst string, maxEdge, quality int) error {
	resize := fmt.Sprintf("%dx%d>", maxEdge, maxEdge)
	cmd := exec.Command("magick", src,
		"-resize", resize,
		"-quality", fmt.Sprintf("%d", quality),
		dst,
	)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
