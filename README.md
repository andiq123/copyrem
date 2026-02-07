# CopyRem

Web app that converts audio to MP3 with ffmpeg (configurable slowdown, 320 kbps, 44.1 kHz).

## Requirements

- **Go** 1.25+ (project uses `toolchain go1.25.7` in go.mod — running `go build` will use or download that version automatically)
- **ffmpeg** — use `bin/ffmpeg` in the repo (macOS) or install: `brew install ffmpeg` (macOS), `apt install ffmpeg` (Linux), or from ffmpeg.org (Windows).

## Build & run

```bash
cd frontend && npm install && npm run build && cd ..
go build -o copyrem ./cmd/server
./copyrem
```

Open http://localhost:8080. Use `-addr :3000` or `PORT=3000` to change the port.

**Dev:** Run `./copyrem`, then `cd frontend && npm run dev` for Vite; UI at http://localhost:5173 with proxy to the API.

## Security

- App and API are same-origin (Go serves the React build). Security headers (CSP, X-Frame-Options, etc.) and rate limiting apply to all routes.
- CORS is only set for dev origins (e.g. localhost:5173). In production, leave `CORS_ORIGINS` unset.
- Errors returned to clients are generic; no stack traces or internal paths. Upload: 80 MB max, extension allowlist, sanitized filenames.

## Deployment

Meta tags, Open Graph, JSON-LD, and `robots.txt`/`sitemap.xml` are in `frontend/`. Replace `https://copyrem.app` in `frontend/index.html` and `frontend/public/` with your base URL. Use HTTPS. When behind a reverse proxy, set `TRUST_PROXY=1` so rate limiting uses the client IP from `X-Forwarded-For`.
