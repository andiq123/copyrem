import { useState, useCallback, useEffect } from 'react'

const DEFAULT_API_INFO = {
  max_upload_mb: 80,
  allowed_extensions: ['.mp3', '.m4a', '.wav', '.flac', '.aac', '.ogg'],
  download_suffix: '_modified.mp3',
}

function formatSize(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function acceptFromInfo(info) {
  return (info?.allowed_extensions || DEFAULT_API_INFO.allowed_extensions).join(',')
}

function downloadFilenameFromResponse(dispositionHeader, originalName, suffix) {
  if (dispositionHeader) {
    const m = dispositionHeader.match(/filename="?([^";]+)"?/)
    if (m) return m[1].trim()
  }
  const base = (originalName || 'audio').replace(/\.[^.]+$/, '') || 'audio'
  return base + (suffix || DEFAULT_API_INFO.download_suffix)
}

export default function App() {
  const [apiInfo, setApiInfo] = useState(DEFAULT_API_INFO)
  const [file, setFile] = useState(null)
  const [loading, setLoading] = useState(false)
  const [status, setStatus] = useState(null)
  const [error, setError] = useState(false)
  const [downloadUrl, setDownloadUrl] = useState(null)
  const [downloadFilenameState, setDownloadFilenameState] = useState(null)

  useEffect(() => {
    fetch('/api/info')
      .then((r) => r.ok ? r.json() : null)
      .then((data) => data && setApiInfo(data))
      .catch(() => {})
  }, [])

  const onFileChange = useCallback((e) => {
    const f = e.target.files?.[0]
    if (f) {
      setFile(f)
      setStatus(null)
      setError(false)
      setDownloadUrl((u) => {
        if (u) URL.revokeObjectURL(u)
        return null
      })
      setDownloadFilenameState(null)
    }
  }, [])

  const onDrop = useCallback((e) => {
    e.preventDefault()
    e.currentTarget.classList.remove('dragover')
    const f = e.dataTransfer.files?.[0]
    if (f) {
      setFile(f)
      setStatus(null)
      setError(false)
        setDownloadUrl((u) => {
          if (u) URL.revokeObjectURL(u)
          return null
        })
        setDownloadFilenameState(null)
    }
  }, [])

  const onDragOver = useCallback((e) => {
    e.preventDefault()
    e.currentTarget.classList.add('dragover')
  }, [])

  const onDragLeave = useCallback((e) => {
    e.currentTarget.classList.remove('dragover')
  }, [])

  const onReset = useCallback(() => {
    setFile(null)
    setStatus(null)
    setError(false)
    setDownloadUrl((u) => {
      if (u) URL.revokeObjectURL(u)
      return null
    })
    setDownloadFilenameState(null)
    const input = document.getElementById('file')
    if (input) input.value = ''
  }, [])

  const canReset = file || status || downloadUrl

  const onSubmit = useCallback(
    async (e) => {
      e.preventDefault()
      if (!file) return

      setLoading(true)
      setStatus(null)
      setError(false)
      setDownloadUrl((u) => {
        if (u) URL.revokeObjectURL(u)
        return null
      })
      setDownloadFilenameState(null)

      const form = new FormData()
      form.append('file', file)

      try {
        const res = await fetch('/convert', { method: 'POST', body: form })
        if (!res.ok) {
          const data = await res.json().catch(() => ({}))
          throw new Error(data.error || res.statusText || 'Conversion failed')
        }
        const blob = await res.blob()
        const disposition = res.headers.get('Content-Disposition')
        const name = downloadFilenameFromResponse(disposition, file.name, apiInfo.download_suffix)
        const url = URL.createObjectURL(blob)
        setDownloadUrl(url)
        setDownloadFilenameState(name)
        setStatus('Your file is ready. It’s been modified so it won’t get detected — same sound, different fingerprint.')
        setError(false)
      } catch (err) {
        setStatus(err.message || 'Something went wrong.')
        setError(true)
      } finally {
        setLoading(false)
      }
    },
    [file, apiInfo.download_suffix]
  )

  return (
    <div className="layout">
      <aside className="aside">
        <h2 className="aside-title">What we do</h2>
        <p className="aside-text">
          We apply a light creative pass (tempo, micro pitch, resample, stereo) so the file is modified enough to avoid detection while still sounding like the original.
        </p>
        <p className="aside-meta">Output: 320 kbps MP3</p>
        <p className="aside-formats">Formats: MP3, M4A, WAV, FLAC, AAC, OGG · max {apiInfo.max_upload_mb} MB</p>
      </aside>

      <div className="center">
        <div className="center-inner">
          <header className="header">
            <h1 className="title">CopyRem</h1>
            <p className="header-tag">Modify your track — same sound, different fingerprint</p>
          </header>

          <main className="main">
            <div className="steps">
              <span className="step step-1">1. Choose file</span>
              <span className="step-divider" aria-hidden>→</span>
              <span className="step step-2">2. Convert</span>
              <span className="step-divider" aria-hidden>→</span>
              <span className="step step-3">3. Download</span>
            </div>

            <form onSubmit={onSubmit} className="upload-form">
            <input
              type="file"
              id="file"
              name="file"
              accept={acceptFromInfo(apiInfo)}
              onChange={onFileChange}
              className="file-input"
              disabled={loading}
            />
            <label
              htmlFor="file"
              className={`dropzone ${loading ? 'dropzone-disabled' : ''}`}
              onDrop={onDrop}
              onDragOver={onDragOver}
              onDragLeave={onDragLeave}
            >
              <span className="dropzone-icon" aria-hidden>↑</span>
              <span className="dropzone-text">
                {file ? file.name : 'Drop file or click'}
              </span>
              {file && (
                <span className="dropzone-meta">{formatSize(file.size)}</span>
              )}
            </label>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={!file || loading}
            >
              {loading ? 'Modifying…' : 'Modify & convert'}
            </button>
          </form>

          {loading && (
            <div className="loading-card" role="status" aria-live="polite">
              <div className="spinner" aria-hidden />
              <p className="loading-title">Modifying your file…</p>
            </div>
          )}

          {status && !loading && (
            <div
              className={`status ${error ? 'error' : 'success'}`}
              role="status"
            >
              {status}
            </div>
          )}

          {downloadUrl && !loading && (
            <div className="result">
              <a
                href={downloadUrl}
                className="btn btn-success"
                download={downloadFilenameState || `audio${apiInfo.download_suffix}`}
              >
                Download
              </a>
            </div>
          )}

            {canReset && !loading && (
              <div className="reset-wrap">
                <button type="button" className="btn btn-ghost" onClick={onReset}>
                  Start again
                </button>
              </div>
            )}
        </main>
        </div>
      </div>
    </div>
  )
}
