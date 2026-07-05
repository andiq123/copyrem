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
        className={`dropzone ${file ? 'has-file' : ''} ${disabled ? 'is-disabled' : ''}`}
        onDrop={onDrop}
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
      >
        <UploadCloud className="dropzone-icon" size={28} strokeWidth={1.5} aria-hidden="true" />
        {file ? (
          <>
            <span className="dropzone-filename">{file.name}</span>
            <span className="dropzone-meta">{formatSize(file.size)}</span>
          </>
        ) : (
          <>
            <span className="dropzone-label">Drop audio here</span>
            <span className="dropzone-hint">MP3, M4A, WAV, FLAC, AAC, OGG</span>
          </>
        )}
      </label>
    </>
  )
}
