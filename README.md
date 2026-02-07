# CopyRem

Modify your audio file so it won't get detected — same sound, different fingerprint.

**Live:** [everyday-deeann-andi3-92f6ccf7.koyeb.app](https://everyday-deeann-andi3-92f6ccf7.koyeb.app/)

Supports MP3, M4A, WAV, FLAC, AAC, OGG. Output: 320 kbps MP3.

## Requirements

- **Go** 1.22+
- **Node.js** 18+
- **ffmpeg** — included in `bin/` (macOS) or install via your package manager

## Quick start

```bash
cd frontend && npm install && npm run build && cd ..
go build -o copyrem .
./copyrem
```

Open [localhost:8080](http://localhost:8080). Set `PORT` to change it.

## Development

```bash
go build -o copyrem . && ./copyrem
cd frontend && npm run dev
```

Backend on `:8080`, Vite on `:5173` with proxy to backend.

## Deployment

Use HTTPS. Set `TRUST_PROXY=1` behind a reverse proxy. Update the canonical URL in `frontend/index.html` to match your domain.
