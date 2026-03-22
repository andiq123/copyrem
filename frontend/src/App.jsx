import { useState, useRef, useEffect } from 'react'
import { Zap, Download, XCircle } from 'lucide-react'
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

  // Dynamic Audio Duration
  const [duration, setDuration] = useState(null)
  
  useEffect(() => {
    if (!file) return
    
    const objectUrl = URL.createObjectURL(file)
    const audio = new Audio(objectUrl)
    
    audio.onloadedmetadata = () => {
      const mins = Math.floor(audio.duration / 60)
      const secs = Math.floor(audio.duration % 60)
      setDuration(`${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`)
    }
    
    return () => URL.revokeObjectURL(objectUrl)
  }, [file])
  
  const displayDuration = file ? (duration || '00:00') : '00:00'

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
        <div className="noise-overlay" />
      </div>

      <div className="app-container" aria-busy={loading} aria-live="polite">

      <Branding />

      <main className="panel">
        <div className="panel-header">
          <div className="panel-title-group">
            <h2 className="panel-title">AUDIO TRANSFORMATION ENGINE</h2>
            <p className="panel-subtitle">Select and process your audio files</p>
          </div>
          <div className="panel-meta">
            <div className="meta-row">
              <span>DURATION:</span>
              <span>{displayDuration}</span>
            </div>
            <div className="meta-row"><span>STATUS:</span> {status ? 'COMPLETED' : (file ? 'READY' : 'STANDBY')}</div>
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
                    <span>Execute Transformation</span>
                  </button>
                )}
              </div>
            </div>
          </form>
        </section>

        {status && !loading && (
          <StatusMessage message={status} isError={error} />
        )}

        {downloadUrl && !loading && (
          <div className="result-card">
            <a
              href={downloadUrl}
              className="btn-download"
              download={downloadName}
              onClick={() => haptic.trigger('success')}
              style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem' }}
            >
              <Download size={20} />
              <span>Download Reconstructed Audio</span>
            </a>
          </div>
        )}

        {canReset && !loading && (
          <div style={{ textAlign: 'center' }}>
            <button type="button" className="btn-ghost" onClick={handleReset}>
              Clear Workspace
            </button>
          </div>
        )}
      </main>

      <footer className="info-footer">
        <span>Engine v2.4</span>
        <span className="dot">•</span>
        <span>{apiInfo?.max_upload_mb ?? 80}MB Limit</span>
      </footer>
      </div>
    </>
  )
}
