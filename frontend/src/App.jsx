import { useState, useRef } from 'react'
import { useWebHaptics } from 'web-haptics/react'
import useConverter from './hooks/useConverter'
import Dropzone from './components/Dropzone'
import ProgressCard from './components/ProgressCard'
import StatusMessage from './components/StatusMessage'

export default function App() {
  const haptic = useWebHaptics()
  const fileInputRef = useRef(null)
  const {
    apiInfo, file, loading, percent, status, error,
    downloadUrl, downloadName, accept, canReset,
    pickFile, submit, reset, cancel,
  } = useConverter()

  const [intensity, setIntensity] = useState(1.0)
  const handleSubmit = (e) => {
    e.preventDefault()
    haptic.trigger([40, 35, 90])
    submit(intensity)
  }

  const handleReset = () => {
    haptic.trigger('medium')
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
        <div className="intensity-control">
          <div className="intensity-header">
            <label htmlFor="intensity-slider" className="intensity-label">Intensity</label>
            <span className="intensity-value">{intensity.toFixed(2)}x</span>
          </div>
          <input
            id="intensity-slider"
            type="range"
            min="0.5"
            max="2.0"
            step="0.05"
            value={intensity}
            onChange={(e) => {
              const val = parseFloat(e.target.value)
              setIntensity(val)
              if (Math.abs(val - 1.0) < 0.01) haptic.trigger('soft')
            }}
            disabled={loading}
            className="intensity-slider"
          />
          <div className="intensity-labels">
            <span>Subtle</span>
            <span>Default</span>
            <span>Heavy</span>
          </div>
        </div>

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
            <button
              type="button"
              className="btn btn-cancel"
              onClick={() => {
                haptic.trigger('warning')
                cancel()
              }}
            >
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
            onClick={() => haptic.trigger('success')}
          >
            <span className="btn-success__icon" aria-hidden>
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="7 10 12 15 17 10" />
                <line x1="12" y1="15" x2="12" y2="3" />
              </svg>
            </span>
            Download your file
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
