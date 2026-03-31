package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const livereloadScript = `<script>
(function() {
    var source = new EventSource('/__livereload');
    source.onmessage = function(event) {
        var data = JSON.parse(event.data);
        if (data.type === 'css') {
            var links = document.querySelectorAll('link[rel="stylesheet"]');
            links.forEach(function(link) {
                var url = new URL(link.href);
                if (url.pathname === data.path) {
                    url.searchParams.set('_lr', Date.now());
                    link.href = url.toString();
                }
            });
        } else if (data.type === 'html') {
            var pagePath = location.pathname;
            if (pagePath.endsWith('/')) pagePath += 'index.html';
            if (data.path === pagePath) location.reload();
        } else {
            location.reload();
        }
    };
})();
</script>`

// sseHub manages connected live-reload clients.
type sseHub struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func newSSEHub() *sseHub {
	return &sseHub{clients: make(map[chan string]struct{})}
}

func (h *sseHub) subscribe() chan string {
	ch := make(chan string, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *sseHub) unsubscribe(ch chan string) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

func (h *sseHub) broadcast(data string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- data:
		default:
			// Drop if client is slow.
		}
	}
}

func runServe(root string) {
	hub := newSSEHub()

	// Initial build.
	fmt.Println("  \033[33m[init]\033[0m building site...")
	if err := buildAll(root); err != nil {
		log.Printf("  \033[31m[error]\033[0m build failed: %v", err)
	}

	// Start file watcher.
	go watchFiles(root, hub)

	// HTTP server.
	mux := http.NewServeMux()
	mux.HandleFunc("/__livereload", func(w http.ResponseWriter, r *http.Request) {
		handleSSE(w, r, hub)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleFile(w, r, root)
	})

	addr := "0.0.0.0:8000"
	fmt.Printf("Serving on http://%s (live-reload enabled)\n", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		fatal("server: %v", err)
	}
}

func handleSSE(w http.ResponseWriter, r *http.Request, hub *sseHub) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := hub.subscribe()
	defer hub.unsubscribe(ch)

	// Send keepalive immediately so the connection is established.
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func handleFile(w http.ResponseWriter, r *http.Request, root string) {
	urlPath := r.URL.Path

	// Map URL path to filesystem.
	fpath := filepath.Join(root, filepath.FromSlash(urlPath))

	// Directory: redirect to trailing slash, then look for index.html.
	info, err := os.Stat(fpath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if info.IsDir() {
		if !strings.HasSuffix(urlPath, "/") {
			http.Redirect(w, r, urlPath+"/", http.StatusMovedPermanently)
			return
		}
		indexPath := filepath.Join(fpath, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			http.NotFound(w, r)
			return
		}
		fpath = indexPath
	}

	// Detect content type.
	ext := filepath.Ext(fpath)
	ctype := mime.TypeByExtension(ext)
	if ctype == "" {
		ctype = "application/octet-stream"
	}

	// No-cache headers for development.
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// For HTML files, inject the live-reload script.
	if strings.HasPrefix(ctype, "text/html") {
		serveHTMLWithLiveReload(w, fpath, ctype)
		return
	}

	w.Header().Set("Content-Type", ctype)
	f, err := os.Open(fpath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func serveHTMLWithLiveReload(w http.ResponseWriter, fpath, ctype string) {
	content, err := os.ReadFile(fpath)
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}

	// Inject live-reload script before </body>.
	injected := strings.Replace(string(content), "</body>", livereloadScript+"</body>", 1)

	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(injected)))
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, injected)
}

// watchFiles monitors src/ and css/ for changes and triggers rebuilds.
func watchFiles(root string, hub *sseHub) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("  \033[31m[error]\033[0m watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Recursively add directories under src/ and css/.
	for _, dir := range []string{
		filepath.Join(root, "src"),
		filepath.Join(root, "css"),
		filepath.Join(root, "js"),
	} {
		addWatchDirs(watcher, dir)
	}

	// Debounce: collect events and rebuild after a quiet period.
	var (
		debounceTimer *time.Timer
		lastEvents    = make(map[string]time.Time)
		mu            sync.Mutex
	)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			path := event.Name
			ext := strings.ToLower(filepath.Ext(path))

			// Skip irrelevant files.
			switch ext {
			case ".md", ".yaml", ".yml", ".html", ".css", ".js":
				// Relevant.
			default:
				continue
			}

			// Debounce: ignore duplicate events within 100ms.
			mu.Lock()
			now := time.Now()
			if last, ok := lastEvents[path]; ok && now.Sub(last) < 100*time.Millisecond {
				mu.Unlock()
				continue
			}
			lastEvents[path] = now
			mu.Unlock()

			// Determine if this is a source file change or a public asset change.
			relPath, _ := filepath.Rel(root, path)
			isSource := strings.HasPrefix(relPath, "src/") || strings.HasPrefix(relPath, "src\\")

			if isSource {
				// Source file changed: rebuild, then notify.
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(150*time.Millisecond, func() {
					fmt.Printf("  \033[36m[watch]\033[0m %s changed, rebuilding...\n", relPath)
					if err := buildAll(root); err != nil {
						log.Printf("  \033[31m[error]\033[0m rebuild failed: %v", err)
						return
					}
					// After rebuild, tell browsers to reload.
					hub.broadcast(`{"type":"reload"}`)
				})
			} else {
				// Public CSS/JS changed directly — notify without rebuild.
				webPath := "/" + filepath.ToSlash(relPath)
				switch ext {
				case ".css":
					fmt.Printf("  \033[36m[reload]\033[0m css: %s\n", webPath)
					hub.broadcast(fmt.Sprintf(`{"type":"css","path":"%s"}`, webPath))
				case ".js":
					fmt.Printf("  \033[36m[reload]\033[0m js: %s\n", webPath)
					hub.broadcast(`{"type":"reload"}`)
				case ".html":
					fmt.Printf("  \033[36m[reload]\033[0m html: %s\n", webPath)
					hub.broadcast(fmt.Sprintf(`{"type":"html","path":"%s"}`, webPath))
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("  \033[31m[error]\033[0m watcher: %v", err)
		}
	}
}

// addWatchDirs recursively adds a directory and all subdirectories to the watcher.
func addWatchDirs(watcher *fsnotify.Watcher, dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})
}
