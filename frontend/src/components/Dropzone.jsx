import { useCallback } from 'react'

function formatSize(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export default function Dropzone({ file, accept, disabled, onFile }) {
  const onDrop = useCallback((e) => {
    e.preventDefault()
    e.currentTarget.classList.remove('dragover')
    onFile(e.dataTransfer.files?.[0])
  }, [onFile])

  const onDragOver = useCallback((e) => {
    e.preventDefault()
    e.currentTarget.classList.add('dragover')
  }, [])

  const onDragLeave = useCallback((e) => e.currentTarget.classList.remove('dragover'), [])

  return (
    <>
      <input
        type="file"
        id="file"
        name="file"
        accept={accept}
        onChange={(e) => onFile(e.target.files?.[0])}
        className="file-input"
        disabled={disabled}
      />
      <label
        htmlFor="file"
        className={`dropzone ${disabled ? 'dropzone-disabled' : ''}`}
        onDrop={onDrop}
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
      >
        <span className="dropzone-icon" aria-hidden>&#8593;</span>
        <span className="dropzone-text">
          {file ? file.name : 'Drop file or click to browse'}
        </span>
        {file && <span className="dropzone-meta">{formatSize(file.size)}</span>}
      </label>
    </>
  )
}
