package main

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	hub := newSSEHub()

	// Initial build.
	fmt.Println("  \033[33m[init]\033[0m building site...")
	if err := buildAll(root); err != nil {
		fmt.Fprintf(os.Stderr, "  \033[31m[error]\033[0m build failed: %v\n", err)
	}

	// Start file watcher with cancellation support.
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		watchFiles(ctx, root, hub)
	}()

	// HTTP server serves from public/.
	publicDir := filepath.Join(root, "public")
	mux := http.NewServeMux()
	mux.HandleFunc("/__livereload", func(w http.ResponseWriter, r *http.Request) {
		handleSSE(w, r, hub)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleFile(w, r, publicDir)
	})

	addr := "0.0.0.0:8000"
	srv := &http.Server{Addr: addr, Handler: mux}

	// Start server in a goroutine.
	srvErr := make(chan error, 1)
	go func() {
		fmt.Printf("Serving on http://%s (live-reload enabled)\n", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			srvErr <- err
		}
		close(srvErr)
	}()

	// Wait for shutdown signal or server error.
	select {
	case <-ctx.Done():
		fmt.Println("\nShutting down...")
	case err := <-srvErr:
		fatal("server: %v", err)
	}

	// Graceful shutdown with 5s timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)

	// Wait for watcher to finish.
	<-watchDone
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
	// Resolve and sanitise the path to prevent traversal.
	cleaned := filepath.Clean(filepath.FromSlash(r.URL.Path))
	fpath := filepath.Join(root, cleaned)
	if !strings.HasPrefix(fpath, root) {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(fpath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if info.IsDir() {
		if !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
			return
		}
		indexPath := filepath.Join(fpath, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			http.NotFound(w, r)
			return
		}
		fpath = indexPath
	}

	ext := filepath.Ext(fpath)
	ctype := mime.TypeByExtension(ext)
	if ctype == "" {
		ctype = "application/octet-stream"
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

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

	injected := strings.Replace(string(content), "</body>", livereloadScript+"</body>", 1)

	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(injected)))
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, injected)
}

// watchFiles monitors src/ and public/ assets for changes. It returns when ctx is cancelled.
func watchFiles(ctx context.Context, root string, hub *sseHub) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  \033[31m[error]\033[0m watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	for _, dir := range []string{
		filepath.Join(root, "src"),
		filepath.Join(root, "public", "stylesheets"),
		filepath.Join(root, "public", "javascripts"),
	} {
		addWatchDirs(watcher, dir)
	}

	var (
		debounceTimer *time.Timer
		lastEvents    = make(map[string]time.Time)
		mu            sync.Mutex
	)

	for {
		select {
		case <-ctx.Done():
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			path := event.Name
			ext := strings.ToLower(filepath.Ext(path))

			switch ext {
			case ".md", ".yaml", ".yml", ".html", ".css", ".js":
			default:
				continue
			}

			// Debounce: ignore duplicate events within 100ms.
			mu.Lock()
			now := time.Now()
			if last, exists := lastEvents[path]; exists && now.Sub(last) < 100*time.Millisecond {
				mu.Unlock()
				continue
			}
			lastEvents[path] = now

			// Prune stale entries to prevent unbounded growth.
			if len(lastEvents) > 200 {
				for k, v := range lastEvents {
					if now.Sub(v) > 10*time.Second {
						delete(lastEvents, k)
					}
				}
			}
			mu.Unlock()

			relPath, _ := filepath.Rel(root, path)
			isSource := strings.HasPrefix(relPath, "src"+string(filepath.Separator))

			if isSource {
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				changedFile := relPath
				debounceTimer = time.AfterFunc(150*time.Millisecond, func() {
					fmt.Printf("  \033[36m[watch]\033[0m %s changed, rebuilding...\n", changedFile)
					if err := buildAll(root); err != nil {
						fmt.Fprintf(os.Stderr, "  \033[31m[error]\033[0m rebuild failed: %v\n", err)
						return
					}
					hub.broadcast(`{"type":"reload"}`)
				})
			} else {
				// Public asset: strip "public/" prefix for the web path.
				publicDir := filepath.Join(root, "public")
				webRel, _ := filepath.Rel(publicDir, path)
				webPath := "/" + filepath.ToSlash(webRel)
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
			fmt.Fprintf(os.Stderr, "  \033[31m[error]\033[0m watcher: %v\n", err)
		}
	}
}

// addWatchDirs recursively adds a directory and all subdirectories to the watcher.
func addWatchDirs(watcher *fsnotify.Watcher, dir string) {
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			watcher.Add(path)
		}
		return nil
	})
}
