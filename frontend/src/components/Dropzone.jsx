import { useCallback } from 'react'
import { useWebHaptics } from 'web-haptics/react'
import { UploadCloud } from 'lucide-react'
import { formatSize } from '../utils/formatSize'

export default function Dropzone({ file, accept, disabled, onFile, inputRef }) {
  const haptic = useWebHaptics()

  const onDrop = useCallback((e) => {
    e.preventDefault()
    e.currentTarget.classList.remove('dragover')
    if (disabled) return
    const f = e.dataTransfer.files?.[0]
    if (f) {
      haptic.trigger('heavy')
      onFile(f)
    }
  }, [onFile, haptic, disabled])

  const onDragOver = useCallback((e) => {
    e.preventDefault()
    if (disabled) return
    e.currentTarget.classList.add('dragover')
  }, [disabled])

  const onDragLeave = useCallback((e) => {
    const related = e.relatedTarget
    if (!related || !e.currentTarget.contains(related)) {
      e.currentTarget.classList.remove('dragover')
    }
  }, [])

  const onInputChange = useCallback((e) => {
    const f = e.target.files?.[0]
    if (f) {
      haptic.trigger('nudge')
      onFile(f)
    }
  }, [onFile, haptic])

  return (
    <>
      <input
        ref={inputRef}
        type="file"
        id="file"
        name="file"
        accept={accept}
        onChange={onInputChange}
        className="file-input"
        disabled={disabled}
        aria-label="Choose audio file"
      />
      <label
        htmlFor="file"
        className={`dropzone-container ${file ? 'is-active' : ''} ${disabled ? 'disabled' : ''}`}
        onDrop={onDrop}
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
      >
        <div className="dropzone-icon">
          <UploadCloud size={44} strokeWidth={1} style={{ filter: 'drop-shadow(0 0 10px var(--accent-glow))', color: 'var(--accent)' }} />
        </div>
        <div className="dropzone-text">
          {file ? (
            <div className="filename-display">{file.name}</div>
          ) : (
            <>
              <div>Drop audio here</div>
              <div className="brand-tag" style={{ marginTop: '4px', fontSize: '0.65rem' }}>or click to browse</div>
            </>
          )}
        </div>
        {file && <div className="brand-tag" style={{ marginTop: '0.5rem', opacity: 0.5 }}>{formatSize(file.size)}</div>}
      </label>
    </>
  )
}
