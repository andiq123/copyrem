import { useState, useRef } from 'react'
import { Zap, Download, CheckCircle2 } from 'lucide-react'
import { useWebHaptics } from 'web-haptics/react'
import useConverter from './hooks/useConverter'
import Dropzone from './components/Dropzone'
import ProgressCard from './components/ProgressCard'
import StatusMessage from './components/StatusMessage'
import Branding from './components/Branding'
import IntensitySlider from './components/IntensitySlider'

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
    // Note: Future backend update needed to accept style, morph, texture
    submit(intensity)
  }

  const handleReset = () => {
    haptic.trigger('medium')
    reset()
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  const isSubmitDisabled = !file || loading

  return (
    <>
      <div className="ambient-bg">
        <div className="orb orb-1" />
        <div className="orb orb-2" />
        <div className="orb orb-3" />
        <div className="orb orb-4" />
      </div>

      <div className="app-container" aria-busy={loading} aria-live="polite">

      <Branding />

      <main className="panel">
        <div className="panel-header">
          <div className="panel-title-group">
            <h2 className="panel-title">Upload your audio</h2>
            <p className="panel-subtitle">Choose a file, set strength, download your MP3</p>
          </div>

        </div>

        <section className="transformation-engine">
          <form onSubmit={handleSubmit} className="engine-grid" noValidate>
            <div className="grid-left">
              <Dropzone
                file={file}
                accept={accept}
                disabled={loading}
                onFile={pickFile}
                inputRef={fileInputRef}
              />
            </div>

            <div className="grid-right">
              <IntensitySlider 
                value={intensity} 
                onChange={setIntensity} 
                disabled={loading} 
              />
              
              <div className="engine-actions">
                {loading ? (
                  <div className="processing-container">
                    <ProgressCard 
                      percent={percent} 
                      onCancel={() => {
                        haptic.trigger('warning')
                        cancel()
                      }}
                    />
                  </div>
                ) : (
                  <button
                    type="submit"
                    className="btn-primary btn-execute"
                    disabled={isSubmitDisabled}
                    style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem' }}
                  >
                    <Zap size={20} />
                    <span>Process and download</span>
                  </button>
                )}
              </div>
            </div>
          </form>
        </section>

        {(status && !downloadUrl) && !loading && (
          <StatusMessage message={status} isError={error} />
        )}

        {downloadUrl && !loading && (
          <div className="result-card">
            <div className="result-header">
              <CheckCircle2 size={20} className="text-accent" />
              <span className="result-text">{status}</span>
            </div>
            <a
              href={downloadUrl}
              className="btn-download"
              download={downloadName}
              onClick={() => haptic.trigger('success')}
            >
              <Download size={20} />
              <span>Download MP3</span>
            </a>
          </div>
        )}

        {canReset && !loading && (
          <div style={{ textAlign: 'center' }}>
            <button type="button" className="btn-ghost" onClick={handleReset}>
              Start over
            </button>
          </div>
        )}
      </main>

      <footer className="info-footer">
        <span>Free, no signup</span>
        <span className="dot">•</span>
        <span>Up to {apiInfo?.max_upload_mb ?? 80}MB per file</span>
      </footer>
      </div>
    </>
  )
}
