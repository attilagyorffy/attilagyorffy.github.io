package main

import (
	"fmt"
	"html"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ogWidth  = 1200
	ogHeight = 630
)

// generateOGImages generates Open Graph social sharing images for all blog
// posts and a default site image. It reads title/subtitle from each post's
// frontmatter, renders an HTML card template, and screenshots it with
// headless Chrome at 1200x630.
func generateOGImages(root string) error {
	srcDir := filepath.Join(root, "src")
	publicDir := filepath.Join(root, "public")
	cardTmplPath := filepath.Join(srcDir, "templates", "og-card.html")
	defaultTmplPath := filepath.Join(srcDir, "templates", "og-card-default.html")

	chrome, err := findChrome()
	if err != nil {
		return err
	}

	// Start a temporary HTTP server so Chrome can load fonts via /fonts/.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting temp listener: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	srv := &http.Server{Handler: http.FileServer(http.Dir(publicDir))}
	go srv.Serve(listener)
	defer srv.Close()

	// Helper: write HTML to a temp file inside public/, screenshot it, remove it.
	renderPath := filepath.Join(publicDir, "_og-render.html")
	defer os.Remove(renderPath)

	screenshot := func(htmlContent []byte, output string) error {
		if err := os.WriteFile(renderPath, htmlContent, 0o644); err != nil {
			return fmt.Errorf("writing render html: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return fmt.Errorf("creating output dir: %w", err)
		}
		url := fmt.Sprintf("http://127.0.0.1:%d/_og-render.html", port)
		cmd := exec.Command(chrome,
			"--headless=new",
			"--screenshot="+output,
			fmt.Sprintf("--window-size=%d,%d", ogWidth, ogHeight),
			"--hide-scrollbars",
			"--disable-gpu",
			"--no-sandbox",
			url,
		)
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Generate default OG image.
	defaultHTML, err := os.ReadFile(defaultTmplPath)
	if err != nil {
		return fmt.Errorf("reading default template: %w", err)
	}
	defaultOut := filepath.Join(publicDir, "images", "og-default.png")
	fmt.Printf("  \033[36m[og]\033[0m images/og-default.png\n")
	if err := screenshot(defaultHTML, defaultOut); err != nil {
		return fmt.Errorf("default og image: %w", err)
	}

	// Generate per-post OG images from frontmatter.
	cardTmpl, err := os.ReadFile(cardTmplPath)
	if err != nil {
		return fmt.Errorf("reading card template: %w", err)
	}
	cardStr := string(cardTmpl)

	md := newMarkdown()
	blogDir := filepath.Join(srcDir, "content", "blog")
	entries, err := os.ReadDir(blogDir)
	if err != nil {
		return fmt.Errorf("reading blog dir: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		mdPath := filepath.Join(blogDir, slug, "index.md")
		if _, err := os.Stat(mdPath); err != nil {
			continue
		}

		post, err := parseBlogPost(mdPath, slug, md)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", slug, err)
		}

		rendered := strings.ReplaceAll(cardStr, "TITLE_PLACEHOLDER", html.EscapeString(post.Title))
		rendered = strings.ReplaceAll(rendered, "TAGLINE_PLACEHOLDER", html.EscapeString(post.Subtitle))

		outPath := filepath.Join(srcDir, "content", "blog", slug, "og.png")
		fmt.Printf("  \033[36m[og]\033[0m %s\n", slug)
		if err := screenshot([]byte(rendered), outPath); err != nil {
			return fmt.Errorf("og for %s: %w", slug, err)
		}
		count++
	}

	fmt.Printf("  \033[32m[og]\033[0m %d post images + default\n", count)
	return nil
}

func findChrome() (string, error) {
	macPath := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	if _, err := os.Stat(macPath); err == nil {
		return macPath, nil
	}
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("chrome not found — install Google Chrome or Chromium")
}
