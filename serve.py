#!/usr/bin/env python3
# /// script
# dependencies = ["watchdog"]
# ///
"""Development server with live-reload via SSE and native filesystem events."""

import json
import os
import queue
import threading
import time
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer

from watchdog.events import FileSystemEventHandler
from watchdog.observers import Observer

WATCHED_EXTENSIONS = {".html", ".css", ".js"}

LIVERELOAD_SCRIPT = b"""<script>
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
</script>
"""


class FileWatcher(FileSystemEventHandler):
    def __init__(self, root_dir):
        self.root = root_dir
        self.clients = []
        self.lock = threading.Lock()
        self._last_event = {}

    def register(self):
        q = queue.Queue()
        with self.lock:
            self.clients.append(q)
        return q

    def unregister(self, q):
        with self.lock:
            try:
                self.clients.remove(q)
            except ValueError:
                pass

    def _notify(self, data):
        with self.lock:
            for q in self.clients:
                q.put(data)

    def _handle_event(self, path):
        if os.path.isdir(path):
            return
        ext = os.path.splitext(path)[1].lower()
        if ext not in WATCHED_EXTENSIONS:
            return

        # Debounce: ignore duplicate events within 100ms
        now = time.monotonic()
        if path in self._last_event and now - self._last_event[path] < 0.1:
            return
        self._last_event[path] = now

        web_path = "/" + os.path.relpath(path, self.root)
        if ext == ".css":
            event_type = "css"
        elif ext == ".js":
            event_type = "js"
        else:
            event_type = "html"

        print(f"  \033[36m[reload]\033[0m {event_type}: {web_path}")
        self._notify(json.dumps({"type": event_type, "path": web_path}))

    def on_modified(self, event):
        self._handle_event(event.src_path)

    def on_created(self, event):
        self._handle_event(event.src_path)

    def on_moved(self, event):
        self._handle_event(event.dest_path)


class LiveReloadHandler(SimpleHTTPRequestHandler):
    watcher = None

    def end_headers(self):
        self.send_header("Cache-Control", "no-cache, no-store, must-revalidate")
        self.send_header("Pragma", "no-cache")
        self.send_header("Expires", "0")
        super().end_headers()

    def do_GET(self):
        if self.path == "/__livereload":
            return self._handle_sse()

        path = self.translate_path(self.path)

        if os.path.isdir(path):
            if not self.path.endswith("/"):
                self.send_response(301)
                self.send_header("Location", self.path + "/")
                self.send_header("Content-Length", "0")
                self.end_headers()
                return
            for index in ("index.html", "index.htm"):
                index_path = os.path.join(path, index)
                if os.path.exists(index_path):
                    path = index_path
                    break
            else:
                return super().do_GET()

        if not os.path.isfile(path):
            self.send_error(404, "File not found")
            return

        ctype = self.guess_type(path)
        if "text/html" in ctype:
            self._serve_html_with_injection(path, ctype)
        else:
            super().do_GET()

    def _serve_html_with_injection(self, path, ctype):
        with open(path, "rb") as f:
            content = f.read()
        content = content.replace(b"</body>", LIVERELOAD_SCRIPT + b"</body>", 1)
        self.send_response(200)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(content)))
        self.end_headers()
        self.wfile.write(content)

    def _handle_sse(self):
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Connection", "keep-alive")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()

        q = self.watcher.register()
        try:
            while True:
                try:
                    data = q.get(timeout=30)
                    self.wfile.write(f"data: {data}\n\n".encode())
                    self.wfile.flush()
                except queue.Empty:
                    self.wfile.write(b": keepalive\n\n")
                    self.wfile.flush()
        except (BrokenPipeError, ConnectionResetError, OSError):
            pass
        finally:
            self.watcher.unregister(q)

    def log_message(self, format, *args):
        if args and "/__livereload" in str(args[0]):
            return
        super().log_message(format, *args)


if __name__ == "__main__":
    root = os.getcwd()

    watcher = FileWatcher(root)
    observer = Observer()
    observer.schedule(watcher, root, recursive=True)
    observer.start()

    LiveReloadHandler.watcher = watcher
    server = ThreadingHTTPServer(("0.0.0.0", 8000), LiveReloadHandler)
    print("Serving on http://0.0.0.0:8000 (live-reload enabled)")

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")
        observer.stop()
    observer.join()
