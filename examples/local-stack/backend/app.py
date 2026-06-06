import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer


HOST = "0.0.0.0"
PORT = 8000


class Handler(BaseHTTPRequestHandler):
    server_version = "PortholeExampleBackend/0.1"

    def _set_headers(self, status=200):
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.end_headers()

    def _write_json(self, payload, status=200):
        body = json.dumps(payload).encode("utf-8")
        self._set_headers(status)
        self.wfile.write(body)

    def do_OPTIONS(self):
        self._set_headers(204)

    def do_GET(self):
        if self.path == "/":
            self._write_json(
                {
                    "service": "backend",
                    "message": "Porthole example backend",
                    "hostname": os.uname().nodename,
                }
            )
            return

        if self.path == "/health":
            self._write_json({"status": "ok"})
            return

        if self.path == "/hello":
            self._write_json(
                {
                    "service": "backend",
                    "message": "Hello from the Python backend behind Traefik",
                    "hostname": os.uname().nodename,
                    "routes": [
                        "https://frontend.localhost",
                        "https://frontend.localhost/api/hello",
                        "https://api.localhost/hello",
                    ],
                }
            )
            return

        self._write_json({"error": "not found", "path": self.path}, status=404)


if __name__ == "__main__":
    HTTPServer((HOST, PORT), Handler).serve_forever()
