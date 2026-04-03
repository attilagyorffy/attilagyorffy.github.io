package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

const siteURL = "https://www.attilagyorffy.com"

// BlogPost holds parsed frontmatter and rendered body for a blog post.
type BlogPost struct {
	Title       string   `yaml:"title"`
	Subtitle    string   `yaml:"subtitle"`
	Date        string   `yaml:"date"`
	Description string   `yaml:"description"`
	Summary     string   `yaml:"summary"`
	Topics      []string `yaml:"topics"`
	Type        string   `yaml:"type"`
	ReadTime    string   `yaml:"read_time"`
	Footer      string   `yaml:"footer"`

	// Derived fields (not from YAML).
	Slug         string
	Body         template.HTML
	DateFormatted string
	DateISO      string
	TopicSlugs   []string
	TopicsJoined string
	HasOGImage   bool
	OGImageURL   string
	URL          string
}

// Page holds parsed frontmatter and rendered body for a standalone page.
type Page struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Layout      string `yaml:"layout"`
	PageClass   string `yaml:"page_class"`
	Footer      string `yaml:"footer"`

	// Derived fields.
	Slug string
	Body template.HTML
	URL  string
}

// YearGroup groups blog posts by year for the listing page.
type YearGroup struct {
	Year  int
	Posts []BlogPost
}

// GalleryData holds the parsed photo gallery YAML.
type GalleryData struct {
	Title        string       `yaml:"title"`
	Slug         string       `yaml:"slug"`
	Date         string       `yaml:"date"`
	DateDisplay  string       `yaml:"date_display"`
	Description  string       `yaml:"description"`
	CountryFlag  string       `yaml:"country_flag"`
	AlbumSummary string       `yaml:"album_summary"`
	Intro        string       `yaml:"intro"`
	CoverImages  []CoverImage `yaml:"cover_images"`
	Photos       []Photo      `yaml:"photos"`
	PhotoCount   int
}

type CoverImage struct {
	Src string `yaml:"src"`
	Alt string `yaml:"alt"`
}

type Photo struct {
	File    string `yaml:"file"`
	W       int    `yaml:"w"`
	H       int    `yaml:"h"`
	Caption string `yaml:"caption"`
}

// PhotoListingData is passed to the photo-listing template.
type PhotoListingData struct {
	Albums []GalleryData
}

// ProjectsData holds the parsed projects.yaml.
type ProjectsData struct {
	Title       string            `yaml:"title"`
	Description string            `yaml:"description"`
	Intro       string            `yaml:"intro"`
	Footer      string            `yaml:"footer"`
	Categories  []ProjectCategory `yaml:"categories"`
}

type ProjectCategory struct {
	Name     string    `yaml:"name"`
	Projects []Project `yaml:"projects"`
}

type Project struct {
	Title   string `yaml:"title"`
	URL     string `yaml:"url"`
	Summary string `yaml:"summary"`
}

// ThoughtsData holds the parsed thoughts.yaml.
type ThoughtsData struct {
	Title       string           `yaml:"title"`
	Description string           `yaml:"description"`
	Intro       string           `yaml:"intro"`
	Footer      string           `yaml:"footer"`
	Sections    []ThoughtSection `yaml:"sections"`
}

type ThoughtSection struct {
	Heading string        `yaml:"heading"`
	Items   []ThoughtItem `yaml:"items"`
}

type ThoughtItem struct {
	Date string `yaml:"date"`
	Text string `yaml:"text"`
}

func main() {
	root := findRoot()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "serve":
			runServe(root)
			return
		case "photos":
			if len(os.Args) < 3 {
				fatal("usage: build photos <album-slug>")
			}
			if err := optimizePhotos(root, os.Args[2]); err != nil {
				fatal("%v", err)
			}
			return
		case "og":
			if err := generateOGImages(root); err != nil {
				fatal("%v", err)
			}
			return
		}
	}

	if err := buildAll(root); err != nil {
		fatal("%v", err)
	}
}

func buildAll(root string) error {
	start := time.Now()
	srcDir := filepath.Join(root, "src")
	outDir := filepath.Join(root, "public")
	tmplDir := filepath.Join(srcDir, "templates")

	tmpl, err := loadTemplates(tmplDir)
	if err != nil {
		return fmt.Errorf("loading templates: %w", err)
	}

	posts, err := buildBlogPosts(srcDir, outDir, tmpl)
	if err != nil {
		return fmt.Errorf("building blog posts: %w", err)
	}

	if err := buildBlogListing(posts, outDir, tmpl); err != nil {
		return fmt.Errorf("building blog listing: %w", err)
	}

	if err := buildPages(srcDir, outDir, tmpl); err != nil {
		return fmt.Errorf("building pages: %w", err)
	}

	if err := buildThoughts(srcDir, outDir, tmpl); err != nil {
		return fmt.Errorf("building thoughts: %w", err)
	}

	if err := buildProjects(srcDir, outDir, tmpl); err != nil {
		return fmt.Errorf("building projects: %w", err)
	}

	if err := buildPhotos(srcDir, outDir, tmpl); err != nil {
		return fmt.Errorf("building photos: %w", err)
	}

	fmt.Printf("  \033[32m[build]\033[0m %d posts + pages in %dms\n", len(posts), time.Since(start).Milliseconds())
	return nil
}

// findRoot returns the repository root (parent of build/).
func findRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		fatal("getwd: %v", err)
	}
	// If we're inside the build/ directory, go up one level.
	if filepath.Base(wd) == "build" {
		return filepath.Dir(wd)
	}
	return wd
}

func loadTemplates(tmplDir string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"lower":      strings.ToLower,
		"slugify":    slugify,
		"join":       strings.Join,
		"safeHTML":   func(s string) template.HTML { return template.HTML(s) },
		"dateFormat": formatDate,
		"topicSlug": func(topic string) string {
			return strings.ToLower(strings.ReplaceAll(topic, " ", "-"))
		},
	}

	tmpl := template.New("").Funcs(funcMap)

	// Parse all template files.
	patterns := []string{
		filepath.Join(tmplDir, "*.html"),
		filepath.Join(tmplDir, "partials", "*.html"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			b, err := os.ReadFile(m)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", m, err)
			}
			name := filepath.Base(m)
			if strings.Contains(m, "partials") {
				name = "partials/" + name
			}
			if _, err := tmpl.New(name).Parse(string(b)); err != nil {
				return nil, fmt.Errorf("parsing %s: %w", m, err)
			}
		}
	}
	return tmpl, nil
}

func buildBlogPosts(srcDir, root string, tmpl *template.Template) ([]BlogPost, error) {
	blogDir := filepath.Join(srcDir, "content", "blog")
	entries, err := os.ReadDir(blogDir)
	if err != nil {
		return nil, err
	}

	md := newMarkdown()
	var posts []BlogPost

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
			return nil, fmt.Errorf("parsing %s: %w", slug, err)
		}

		// Check for og.png.
		ogSrc := filepath.Join(blogDir, slug, "og.png")
		ogDst := filepath.Join(root, "blog", slug, "og.png")
		if _, err := os.Stat(ogSrc); err == nil {
			post.HasOGImage = true
			post.OGImageURL = siteURL + "/blog/" + slug + "/og.png"
			if err := copyFile(ogSrc, ogDst); err != nil {
				return nil, fmt.Errorf("copying og.png for %s: %w", slug, err)
			}
		} else {
			post.OGImageURL = siteURL + "/images/og-default.png"
		}

		// Render to HTML.
		outDir := filepath.Join(root, "blog", slug)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating dir %s: %w", outDir, err)
		}
		outPath := filepath.Join(outDir, "index.html")

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "blog-post.html", post); err != nil {
			return nil, fmt.Errorf("rendering %s: %w", slug, err)
		}
		if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
			return nil, fmt.Errorf("writing %s: %w", outPath, err)
		}

		posts = append(posts, post)
	}

	// Sort by date descending.
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].DateISO > posts[j].DateISO
	})

	return posts, nil
}

func parseBlogPost(mdPath, slug string, md goldmark.Markdown) (BlogPost, error) {
	raw, err := os.ReadFile(mdPath)
	if err != nil {
		return BlogPost{}, err
	}

	frontmatter, body, err := splitFrontmatter(raw)
	if err != nil {
		return BlogPost{}, err
	}

	var post BlogPost
	if err := yaml.Unmarshal(frontmatter, &post); err != nil {
		return BlogPost{}, fmt.Errorf("yaml: %w", err)
	}

	// Render markdown body to HTML.
	var buf bytes.Buffer
	if err := md.Convert(body, &buf); err != nil {
		return BlogPost{}, fmt.Errorf("markdown: %w", err)
	}

	rendered := buf.String()

	// Add target/rel to external links.
	rendered = addExternalLinkAttrs(rendered)

	post.Slug = slug
	post.Body = template.HTML(rendered)
	post.URL = "/blog/" + slug + "/"

	// Render footer markdown to HTML (strip wrapping <p> tags).
	if post.Footer != "" {
		var fbuf bytes.Buffer
		md.Convert([]byte(post.Footer), &fbuf)
		footer := addExternalLinkAttrs(fbuf.String())
		// Strip wrapping <p>...</p> since the template already wraps in <p>.
		footer = strings.TrimSpace(footer)
		footer = strings.TrimPrefix(footer, "<p>")
		footer = strings.TrimSuffix(footer, "</p>")
		post.Footer = footer
	}

	// Derive date fields.
	post.DateISO = post.Date
	if t, err := time.Parse("2006-01-02", post.Date); err == nil {
		post.DateFormatted = t.Format("Jan 2, 2006")
	}

	// Derive topic slugs.
	for _, t := range post.Topics {
		post.TopicSlugs = append(post.TopicSlugs, slugify(t))
	}
	post.TopicsJoined = strings.Join(post.TopicSlugs, " ")

	return post, nil
}

func buildBlogListing(posts []BlogPost, root string, tmpl *template.Template) error {
	// Group by year.
	yearMap := map[int][]BlogPost{}
	for _, p := range posts {
		if t, err := time.Parse("2006-01-02", p.DateISO); err == nil {
			yearMap[t.Year()] = append(yearMap[t.Year()], p)
		}
	}

	var years []int
	for y := range yearMap {
		years = append(years, y)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))

	var groups []YearGroup
	for _, y := range years {
		groups = append(groups, YearGroup{Year: y, Posts: yearMap[y]})
	}

	// Collect unique topics for filter buttons.
	topicSet := map[string]bool{}
	for _, p := range posts {
		for _, t := range p.Topics {
			topicSet[t] = true
		}
	}
	var allTopics []string
	for t := range topicSet {
		allTopics = append(allTopics, t)
	}
	sort.Strings(allTopics)

	data := struct {
		YearGroups []YearGroup
		AllTopics  []string
	}{
		YearGroups: groups,
		AllTopics:  allTopics,
	}

	outPath := filepath.Join(root, "blog", "index.html")
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "blog-listing.html", data); err != nil {
		return fmt.Errorf("rendering listing: %w", err)
	}
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func buildPages(srcDir, root string, tmpl *template.Template) error {
	pagesDir := filepath.Join(srcDir, "content", "pages")
	entries, err := os.ReadDir(pagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	md := newMarkdown()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		slug := strings.TrimSuffix(entry.Name(), ".md")
		mdPath := filepath.Join(pagesDir, entry.Name())

		raw, err := os.ReadFile(mdPath)
		if err != nil {
			return err
		}

		fm, body, err := splitFrontmatter(raw)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", slug, err)
		}

		var page Page
		if err := yaml.Unmarshal(fm, &page); err != nil {
			return fmt.Errorf("yaml %s: %w", slug, err)
		}

		var buf bytes.Buffer
		if err := md.Convert(body, &buf); err != nil {
			return fmt.Errorf("markdown %s: %w", slug, err)
		}

		page.Slug = slug
		page.Body = template.HTML(buf.String())
		page.URL = "/" + slug + "/"

		// Render footer markdown to HTML (strip wrapping <p> tags).
		if page.Footer != "" {
			var fbuf bytes.Buffer
			md.Convert([]byte(page.Footer), &fbuf)
			footer := addExternalLinkAttrs(fbuf.String())
			footer = strings.TrimSpace(footer)
			footer = strings.TrimPrefix(footer, "<p>")
			footer = strings.TrimSuffix(footer, "</p>")
			page.Footer = footer
		}

		outDir := filepath.Join(root, slug)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("creating dir %s: %w", outDir, err)
		}
		outPath := filepath.Join(outDir, "index.html")

		var out bytes.Buffer
		if err := tmpl.ExecuteTemplate(&out, "page.html", page); err != nil {
			return fmt.Errorf("rendering %s: %w", slug, err)
		}
		if err := os.WriteFile(outPath, out.Bytes(), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
	}

	return nil
}

func buildPhotos(srcDir, root string, tmpl *template.Template) error {
	// Photos live at src/photos/<slug>/<slug>.yaml alongside raw exports.
	photosDir := filepath.Join(srcDir, "photos")
	entries, err := os.ReadDir(photosDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var albums []GalleryData

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		yamlPath := filepath.Join(photosDir, slug, slug+".yaml")
		raw, err := os.ReadFile(yamlPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		var gallery GalleryData
		if err := yaml.Unmarshal(raw, &gallery); err != nil {
			return fmt.Errorf("yaml %s: %w", entry.Name(), err)
		}
		gallery.PhotoCount = len(gallery.Photos)

		// Render individual gallery page.
		outDir := filepath.Join(root, "photos", gallery.Slug)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("creating dir %s: %w", outDir, err)
		}
		outPath := filepath.Join(outDir, "index.html")

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "photo-gallery.html", gallery); err != nil {
			return fmt.Errorf("rendering gallery %s: %w", gallery.Slug, err)
		}
		if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}

		albums = append(albums, gallery)
	}

	// Sort albums by date descending.
	sort.Slice(albums, func(i, j int) bool {
		return albums[i].Date > albums[j].Date
	})

	// Render photo listing page.
	listData := PhotoListingData{Albums: albums}
	outPath := filepath.Join(root, "photos", "index.html")

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "photo-listing.html", listData); err != nil {
		return fmt.Errorf("rendering photo listing: %w", err)
	}
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func buildProjects(srcDir, root string, tmpl *template.Template) error {
	dataPath := filepath.Join(srcDir, "content", "projects.yaml")
	raw, err := os.ReadFile(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var data ProjectsData
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("yaml: %w", err)
	}

	outDir := filepath.Join(root, "projects")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating dir %s: %w", outDir, err)
	}
	outPath := filepath.Join(outDir, "index.html")

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "projects.html", data); err != nil {
		return fmt.Errorf("rendering projects: %w", err)
	}
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func buildThoughts(srcDir, root string, tmpl *template.Template) error {
	dataPath := filepath.Join(srcDir, "content", "thoughts.yaml")
	raw, err := os.ReadFile(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var data ThoughtsData
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("yaml: %w", err)
	}

	outDir := filepath.Join(root, "thoughts")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating dir %s: %w", outDir, err)
	}
	outPath := filepath.Join(outDir, "index.html")

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "thoughts.html", data); err != nil {
		return fmt.Errorf("rendering thoughts: %w", err)
	}
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

// --- Helpers ---

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.Footnote,
			extension.Strikethrough,
			extension.Typographer,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML passthrough (<mark>, <ul class="takeaway">, etc.).
		),
	)
}

func splitFrontmatter(data []byte) ([]byte, []byte, error) {
	const sep = "---"
	s := string(data)
	s = strings.TrimLeft(s, "\n\r\t ")

	if !strings.HasPrefix(s, sep) {
		return nil, []byte(s), nil
	}

	rest := s[len(sep):]
	idx := strings.Index(rest, "\n"+sep)
	if idx < 0 {
		return nil, []byte(s), nil
	}

	fm := rest[:idx]
	body := rest[idx+len("\n"+sep):]
	body = strings.TrimLeft(body, "\n\r")

	return []byte(fm), []byte(body), nil
}

// addExternalLinkAttrs adds target="_blank" rel="noopener noreferrer" to
// links that point outside the site.
var linkRe = regexp.MustCompile(`<a\s+href="(https?://[^"]*)"`)

func addExternalLinkAttrs(s string) string {
	return linkRe.ReplaceAllStringFunc(s, func(match string) string {
		if strings.Contains(match, `target=`) {
			return match // already has target
		}
		return match + ` target="_blank" rel="noopener noreferrer"`
	})
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "")
	return s
}

func formatDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("Jan 2, 2006")
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating dir for %s: %w", dst, err)
	}
	return os.WriteFile(dst, data, 0o644)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "build: "+format+"\n", args...)
	os.Exit(1)
}
