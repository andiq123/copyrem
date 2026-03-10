# CopyRem

**Same sound, different fingerprint.** Upload an audio file and get a modified version that keeps the sound but changes the fingerprint—so it won’t get detected by content ID or matching systems.

**Live:** [everyday-deeann-andi3-92f6ccf7.koyeb.app](https://everyday-deeann-andi3-92f6ccf7.koyeb.app/)

Accepts MP3, M4A, WAV, FLAC, AAC, OGG. Output: 320kbps MP3. Free, no signup.

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

## Troubleshooting

**`npm warn Unknown env config "devdir"`** — Cursor (or another tool) sets `npm_config_devdir` for node-gyp; npm doesn’t recognize it. Safe to ignore, or clear it before running npm:

```bash
unset npm_config_devdir
cd frontend && npm install && npm run build
```
