import { useState, useRef, useEffect } from 'react'
import { Download } from 'lucide-react'
import { useWebHaptics } from 'web-haptics/react'
import useConverter, { MAX_UPLOAD_MB } from './hooks/useConverter'
import { previewUrlForFile } from './utils/formatSize'
import Dropzone from './components/Dropzone'
import ProgressCard from './components/ProgressCard'
import StatusMessage from './components/StatusMessage'
import Branding from './components/Branding'
import IntensitySlider from './components/IntensitySlider'
import AudioPreview from './components/AudioPreview'

export default function App() {
  const haptic = useWebHaptics()
  const fileInputRef = useRef(null)
  const [originalUrl, setOriginalUrl] = useState(null)

  const {
    file, loading, percent, status, error,
    downloadUrl, downloadName, accept, canReset,
    pickFile, submit, reset, cancel,
  } = useConverter()

  const [intensity, setIntensity] = useState(1.0)

  useEffect(() => {
    if (!file) {
      setOriginalUrl(null)
      return
    }
    const url = previewUrlForFile(file)
    setOriginalUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [file])

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

          {originalUrl && !downloadUrl && !loading && (
            <AudioPreview src={originalUrl} title={file.name} badge="Original" />
          )}

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
                {downloadUrl ? 'Process again' : 'Process audio'}
              </button>
            )}
          </div>
        </form>

        {(status && !downloadUrl) && !loading && (
          <StatusMessage message={status} isError={error} />
        )}

        {downloadUrl && !loading && (
          <div className="preview-stack">
            <p className="preview-heading">Listen before you download</p>
            <AudioPreview
              src={downloadUrl}
              title={downloadName}
              badge="Remixed"
              active
            />
            {originalUrl && (
              <AudioPreview src={originalUrl} title={file.name} badge="Original" />
            )}
            <a
              href={downloadUrl}
              className="btn-download"
              download={downloadName}
              onClick={() => haptic.trigger('success')}
            >
              <Download size={18} aria-hidden="true" />
              Download MP3
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
        <span aria-hidden="true">·</span>
        <span>Up to {MAX_UPLOAD_MB}MB per file</span>
      </footer>
    </div>
  )
}
