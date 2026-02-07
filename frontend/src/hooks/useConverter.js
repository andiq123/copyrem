import { useState, useCallback, useEffect, useRef } from 'react'

const DEFAULT_API_INFO = {
  max_upload_mb: 80,
  allowed_extensions: ['.mp3', '.m4a', '.wav', '.flac', '.aac', '.ogg'],
  download_suffix: '_modified.mp3',
}

export default function useConverter() {
  const [apiInfo, setApiInfo] = useState(DEFAULT_API_INFO)
  const [file, setFile] = useState(null)
  const [loading, setLoading] = useState(false)
  const [percent, setPercent] = useState(0)
  const [status, setStatus] = useState(null)
  const [error, setError] = useState(false)
  const [downloadUrl, setDownloadUrl] = useState(null)
  const [downloadName, setDownloadName] = useState(null)
  const esRef = useRef(null)

  useEffect(() => {
    fetch('/api/info')
      .then((r) => r.ok ? r.json() : null)
      .then((data) => data && setApiInfo(data))
      .catch(() => {})
  }, [])

  const closeES = useCallback(() => {
    if (esRef.current) {
      esRef.current.close()
      esRef.current = null
    }
  }, [])

  const clearState = useCallback(() => {
    setStatus(null)
    setError(false)
    setPercent(0)
    setDownloadUrl((u) => {
      if (u) URL.revokeObjectURL(u)
      return null
    })
    setDownloadName(null)
  }, [])

  const fail = useCallback((msg) => {
    closeES()
    setStatus(msg)
    setError(true)
    setLoading(false)
  }, [closeES])

  const pickFile = useCallback((f) => {
    if (!f) return
    setFile(f)
    clearState()
  }, [clearState])

  const reset = useCallback(() => {
    closeES()
    setFile(null)
    setLoading(false)
    clearState()
  }, [closeES, clearState])

  const submit = useCallback(async () => {
    if (!file) return

    closeES()
    setLoading(true)
    clearState()

    const form = new FormData()
    form.append('file', file)

    try {
      const res = await fetch('/convert', { method: 'POST', body: form })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || res.statusText || 'Conversion failed')
      }
      const { job_id } = await res.json()

      const es = new EventSource(`/convert/progress/${job_id}`)
      esRef.current = es

      es.onmessage = async (event) => {
        const msg = JSON.parse(event.data)
        setPercent(msg.percent || 0)

        if (msg.error) return fail(msg.error)

        if (msg.done) {
          closeES()
          setPercent(100)
          try {
            const dl = await fetch(`/convert/download/${job_id}`)
            if (!dl.ok) throw new Error('Download failed')
            const blob = await dl.blob()
            const disp = dl.headers.get('Content-Disposition')
            const match = disp?.match(/filename="?([^";]+)"?/)
            setDownloadUrl(URL.createObjectURL(blob))
            setDownloadName(match?.[1]?.trim() || `audio${apiInfo.download_suffix}`)
            setStatus('Your file is ready \u2014 same sound, different fingerprint.')
            setLoading(false)
          } catch {
            fail('Download failed.')
          }
        }
      }

      es.onerror = () => fail('Connection lost. Please try again.')
    } catch (err) {
      fail(err.message || 'Something went wrong.')
    }
  }, [file, apiInfo.download_suffix, closeES, clearState, fail])

  const accept = (apiInfo?.allowed_extensions || DEFAULT_API_INFO.allowed_extensions).join(',')

  return {
    apiInfo, file, loading, percent, status, error,
    downloadUrl, downloadName, accept,
    pickFile, submit, reset,
    canReset: !!(file || status || downloadUrl),
  }
}
