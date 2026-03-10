import { useCallback } from 'react'
import { formatSize } from '../utils/formatSize'

export default function Dropzone({ file, accept, disabled, onFile, inputRef }) {
  const onDrop = useCallback((e) => {
    e.preventDefault()
    e.currentTarget.classList.remove('dragover')
    onFile(e.dataTransfer.files?.[0])
  }, [onFile])

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

  return (
    <>
      <input
        ref={inputRef}
        type="file"
        id="file"
        name="file"
        accept={accept}
        onChange={(e) => onFile(e.target.files?.[0])}
        className="file-input"
        disabled={disabled}
        aria-label="Choose audio file"
      />
      <label
        htmlFor="file"
        className={`dropzone ${disabled ? 'dropzone-disabled' : ''}`}
        onDrop={onDrop}
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
      >
        <span className="dropzone-icon" aria-hidden>&#8593;</span>
        <span className="dropzone-text" title={file?.name}>
          {file ? file.name : 'Drop file or click to browse'}
        </span>
        {file && <span className="dropzone-meta">{formatSize(file.size)}</span>}
      </label>
    </>
  )
}
