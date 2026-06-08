# SPDX-License-Identifier: Apache-2.0
# store-api mock — product recommendation endpoint for yaagents examples/store demo.
# Stdlib only: no pip installs required; keeps the example self-contained.
#
# Endpoints:
#   GET  /healthz                           -> 200 {"status":"ok"}
#   POST /products/{id}/recommendations     -> 200 {"recommendations":[...]}
#
# The gateway proxies POST /products/{productId}/recommendations here and
# passes through the X-Actor-Subject header injected from the JWT sub claim
# (in a production system this would be a verified identity; here it is
# echoed in the response for demo visibility).

import http.server
import json
import os
import re
import sys

_RECOMMENDATIONS_BY_PRODUCT = {
    "p-1": [
        {"product_id": "p-2", "score": 0.95, "reason": "frequently bought together"},
        {"product_id": "p-3", "score": 0.88, "reason": "customers also viewed"},
        {"product_id": "p-4", "score": 0.72, "reason": "trending in your category"},
        {"product_id": "p-5", "score": 0.65, "reason": "based on your history"},
    ],
}
_DEFAULT_RECS = [
    {"product_id": "p-1", "score": 0.90, "reason": "top rated"},
    {"product_id": "p-2", "score": 0.82, "reason": "popular this week"},
    {"product_id": "p-3", "score": 0.70, "reason": "new arrival"},
]

_PATH_RE = re.compile(r"^/products/([^/]+)/recommendations$")


class _Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):  # quiet default logger
        sys.stdout.write(f"store-api {self.address_string()} {fmt % args}\n")
        sys.stdout.flush()

    def _send_json(self, code: int, body: dict) -> None:
        payload = json.dumps(body).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def do_GET(self) -> None:
        if self.path == "/healthz":
            self._send_json(200, {"status": "ok"})
        else:
            self._send_json(404, {"error": "not_found"})

    def do_POST(self) -> None:
        m = _PATH_RE.match(self.path)
        if not m:
            self._send_json(404, {"error": "not_found"})
            return

        product_id = m.group(1)
        content_len = int(self.headers.get("Content-Length", 0))
        raw = self.rfile.read(content_len) if content_len else b"{}"
        try:
            req_body = json.loads(raw)
        except json.JSONDecodeError:
            req_body = {}

        limit = int(req_body.get("limit", 3))
        exclude = req_body.get("exclude_purchased", False)
        actor = self.headers.get("X-Actor-Subject", "anonymous")

        all_recs = _RECOMMENDATIONS_BY_PRODUCT.get(product_id, _DEFAULT_RECS)
        recs = all_recs[:limit]

        self._send_json(200, {
            "product_id": product_id,
            "actor": actor,
            "exclude_purchased": exclude,
            "recommendations": recs,
        })


def main() -> None:
    port = int(os.environ.get("PORT", "8125"))
    server = http.server.HTTPServer(("0.0.0.0", port), _Handler)
    sys.stdout.write(f"store-api listening on :{port}\n")
    sys.stdout.flush()
    server.serve_forever()


if __name__ == "__main__":
    main()
