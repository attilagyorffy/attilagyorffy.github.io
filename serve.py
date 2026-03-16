#!/usr/bin/env python3
"""Development server with cache-busting headers."""

from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer


class NoCacheHandler(SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header("Cache-Control", "no-cache, no-store, must-revalidate")
        self.send_header("Pragma", "no-cache")
        self.send_header("Expires", "0")
        super().end_headers()


if __name__ == "__main__":
    server = ThreadingHTTPServer(("localhost", 8000), NoCacheHandler)
    print("Serving on http://localhost:8000 (no-cache)")
    server.serve_forever()
