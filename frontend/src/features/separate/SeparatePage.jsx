import useSeparate from './useSeparate'
import Dropzone from '../../components/Dropzone'
import ProgressCard from '../../components/ProgressCard'
import StatusMessage from '../../components/StatusMessage'

export default function SeparatePage() {
  const {
    apiInfo, file, loading, percent, status, error,
    downloadVocalsUrl, downloadVocalsName,
    downloadInstrumentalUrl, downloadInstrumentalName,
    accept, canReset,
    pickFile, submit, reset, cancel,
  } = useSeparate()

  const onSubmit = (e) => {
    e.preventDefault()
    submit()
  }

  const onReset = () => {
    reset()
    const input = document.getElementById('file')
    if (input) input.value = ''
  }

  return (
    <div className="widget">
      <div className="brand">
        <span className="brand-name">CopyRem</span>
        <span className="brand-tag">Separate vocals and instrumental</span>
      </div>

      <form onSubmit={onSubmit} className="upload-form">
        <Dropzone file={file} accept={accept} disabled={loading} onFile={pickFile} />
        <button type="submit" className="btn btn-primary" disabled={!file || loading}>
          {loading ? 'Processing\u2026' : 'Separate'}
        </button>
      </form>

      {loading && (
        <>
          <ProgressCard percent={percent} />
          <div className="cancel-wrap">
            <button type="button" className="btn btn-cancel" onClick={cancel}>
              Cancel
            </button>
          </div>
        </>
      )}

      {status && !loading && <StatusMessage message={status} isError={error} />}

      {(downloadVocalsUrl || downloadInstrumentalUrl) && !loading && (
        <div className="result">
          <a
            href={downloadVocalsUrl}
            className="btn btn-success"
            download={downloadVocalsName}
          >
            Download vocals
          </a>
          <a
            href={downloadInstrumentalUrl}
            className="btn btn-success"
            download={downloadInstrumentalName}
          >
            Download instrumental
          </a>
        </div>
      )}

      {canReset && !loading && (
        <div className="reset-wrap">
          <button type="button" className="btn btn-ghost" onClick={onReset}>
            Start over
          </button>
        </div>
      )}

      <div className="widget-footer">
        MP3 &middot; M4A &middot; WAV &middot; FLAC &middot; AAC &middot; OGG &middot; max {apiInfo.max_upload_mb} MB
      </div>
    </div>
  )
}
