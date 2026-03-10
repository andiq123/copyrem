import { useCallback } from 'react'
import { useWebHaptics } from 'web-haptics/react'
import { formatSize } from '../utils/formatSize'

export default function Dropzone({ file, accept, disabled, onFile, inputRef }) {
  const haptic = useWebHaptics()

  const onDrop = useCallback((e) => {
    e.preventDefault()
    e.currentTarget.classList.remove('dragover')
    const f = e.dataTransfer.files?.[0]
    if (f) {
      haptic.trigger('heavy')
      onFile(f)
    }
  }, [onFile, haptic])

  const onDragOver = useCallback((e) => {
    e.preventDefault()
    e.currentTarget.classList.add('dragover')
  }, [])

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
        className={`dropzone ${file ? 'dropzone-has-file' : ''} ${disabled ? 'dropzone-disabled' : ''}`}
        onDrop={onDrop}
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
      >
        <span className="dropzone-icon" aria-hidden>
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="17 8 12 3 7 8" />
            <line x1="12" y1="3" x2="12" y2="15" />
          </svg>
        </span>
        <span className="dropzone-text" title={file?.name}>
          {file ? file.name : 'Drop an audio file or click to choose'}
        </span>
        {file && <span className="dropzone-meta">{formatSize(file.size)}</span>}
      </label>
    </>
  )
}
