import { useRef } from 'react'
import useConverter from './hooks/useConverter'
import Dropzone from './components/Dropzone'
import ProgressCard from './components/ProgressCard'
import StatusMessage from './components/StatusMessage'

export default function App() {
  const fileInputRef = useRef(null)
  const {
    apiInfo, file, loading, percent, status, error,
    downloadUrl, downloadName, accept, canReset,
    pickFile, submit, reset, cancel,
  } = useConverter()

  const handleSubmit = (e) => {
    e.preventDefault()
    submit()
  }

  const handleReset = () => {
    reset()
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  return (
    <main className="widget" aria-busy={loading} aria-live="polite">
      <header className="brand">
        <h1 className="brand-name">CopyRem</h1>
        <p className="brand-tag">Same sound, different fingerprint</p>
      </header>

      <form onSubmit={handleSubmit} className="upload-form" noValidate>
        <Dropzone
          file={file}
          accept={accept}
          disabled={loading}
          onFile={pickFile}
          inputRef={fileInputRef}
        />
        <button
          type="submit"
          className="btn btn-primary"
          disabled={!file || loading}
          aria-busy={loading}
        >
          {loading ? 'Processing…' : 'Convert'}
        </button>
      </form>

      {loading && (
        <>
          <ProgressCard percent={percent} />
          <div className="cancel-wrap">
            <button type="button" className="btn btn-cancel" onClick={cancel}>
              Cancel
            </button>
          </div>
        </>
      )}

      {status && !loading && <StatusMessage message={status} isError={error} />}

      {downloadUrl && !loading && (
        <div className="result">
          <a
            href={downloadUrl}
            className="btn btn-success"
            download={downloadName}
          >
            Download
          </a>
        </div>
      )}

      {canReset && !loading && (
        <div className="reset-wrap">
          <button type="button" className="btn btn-ghost" onClick={handleReset}>
            Start over
          </button>
        </div>
      )}

      <footer className="widget-footer">
        MP3 &middot; M4A &middot; WAV &middot; FLAC &middot; AAC &middot; OGG &middot; max {apiInfo?.max_upload_mb ?? 80}&nbsp;MB
      </footer>
    </main>
  )
}
