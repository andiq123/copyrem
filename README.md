# CopyRem

Web app that converts audio to MP3 with ffmpeg (configurable slowdown, 320 kbps, 44.1 kHz).

## Requirements

- **Go** 1.22+
- **ffmpeg** — included in `bin/` (macOS) or install via your package manager.

## Build & run

```bash
cd frontend && npm install && npm run build && cd ..
go build -o copyrem .
./copyrem
```

Open http://localhost:8080. Use `PORT=3000` to change the port.

**Dev:** Run `./copyrem`, then `cd frontend && npm run dev` for Vite at http://localhost:5173.

## Deployment

Replace `https://copyrem.app` in `frontend/index.html` and `frontend/public/` with your base URL. Use HTTPS. Set `TRUST_PROXY=1` when behind a reverse proxy.
