import { useState, useRef } from 'react'
import { Download, CheckCircle2 } from 'lucide-react'
import { useWebHaptics } from 'web-haptics/react'
import useConverter, { MAX_UPLOAD_MB } from './hooks/useConverter'
import Dropzone from './components/Dropzone'
import ProgressCard from './components/ProgressCard'
import StatusMessage from './components/StatusMessage'
import Branding from './components/Branding'
import IntensitySlider from './components/IntensitySlider'

export default function App() {
  const haptic = useWebHaptics()
  const fileInputRef = useRef(null)

  const {
    file, loading, percent, status, error,
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
    <div className="app-container" aria-busy={loading} aria-live="polite">
        <Branding />

        <main className="panel">
          <form onSubmit={handleSubmit} className="panel-stack" noValidate>
            <Dropzone
              file={file}
              accept={accept}
              disabled={loading}
              onFile={pickFile}
              inputRef={fileInputRef}
            />

            <IntensitySlider
              value={intensity}
              onChange={setIntensity}
              disabled={loading}
            />

            <div className="panel-actions">
              {loading ? (
                <ProgressCard
                  percent={percent}
                  onCancel={() => {
                    haptic.trigger('warning')
                    cancel()
                  }}
                />
              ) : (
                <button type="submit" className="btn-primary" disabled={!file}>
                  Process and download
                </button>
              )}
            </div>
          </form>

          {(status && !downloadUrl) && !loading && (
            <StatusMessage message={status} isError={error} />
          )}

          {downloadUrl && !loading && (
            <div className="result-card">
              <div className="result-header">
                <CheckCircle2 size={20} className="text-accent" aria-hidden="true" />
                <span className="result-text">{status}</span>
              </div>
              <a
                href={downloadUrl}
                className="btn-download"
                download={downloadName}
                onClick={() => haptic.trigger('success')}
              >
                <Download size={20} aria-hidden="true" />
                <span>Download MP3</span>
              </a>
            </div>
          )}

          {canReset && !loading && (
            <div className="panel-footer">
              <button type="button" className="btn-ghost" onClick={handleReset}>
                Start over
              </button>
            </div>
          )}
        </main>

        <footer className="info-footer">
          <span>Free, no signup</span>
          <span className="dot" aria-hidden="true">·</span>
          <span>Up to {MAX_UPLOAD_MB}MB per file</span>
        </footer>
    </div>
  )
}
