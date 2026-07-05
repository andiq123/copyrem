export function formatSize(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function formatTime(secs) {
  if (!secs || !Number.isFinite(secs)) return '0:00'
  const m = Math.floor(secs / 60)
  const s = Math.floor(secs % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

const MIME = { mp3: 'audio/mpeg', m4a: 'audio/mp4', wav: 'audio/wav', flac: 'audio/flac', aac: 'audio/aac', ogg: 'audio/ogg' }

export function previewUrlForFile(file) {
  if (file.type) return URL.createObjectURL(file)
  const ext = file.name.split('.').pop()?.toLowerCase()
  return URL.createObjectURL(new Blob([file], { type: MIME[ext] || 'audio/mpeg' }))
}
