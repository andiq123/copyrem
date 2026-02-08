import { useState, useCallback, useEffect, useRef } from 'react'

const DEFAULT_API_INFO = {
  max_upload_mb: 80,
  allowed_extensions: ['.mp3', '.m4a', '.wav', '.flac', '.aac', '.ogg'],
  download_suffix: '_modified.mp3',
}

export default function useConvert() {
  const [apiInfo, setApiInfo] = useState(DEFAULT_API_INFO)
  const [file, setFile] = useState(null)
  const [loading, setLoading] = useState(false)
  const [percent, setPercent] = useState(0)
  const [status, setStatus] = useState(null)
  const [error, setError] = useState(false)
  const [downloadUrl, setDownloadUrl] = useState(null)
  const [downloadName, setDownloadName] = useState(null)
  const esRef = useRef(null)
  const jobIdRef = useRef(null)

  useEffect(() => {
    fetch('/api/info')
      .then((r) => r.ok ? r.json() : null)
      .then((data) => data && setApiInfo(data))
      .catch(() => {})
  }, [])

  useEffect(() => {
    const onUnload = () => {
      if (jobIdRef.current) {
        navigator.sendBeacon(`/convert/cancel/${jobIdRef.current}`)
      }
    }
    window.addEventListener('beforeunload', onUnload)
    return () => window.removeEventListener('beforeunload', onUnload)
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

  const stopJob = useCallback(() => {
    if (jobIdRef.current) {
      fetch(`/convert/cancel/${jobIdRef.current}`, { method: 'POST' }).catch(() => {})
    }
    closeES()
    jobIdRef.current = null
  }, [closeES])

  const fail = useCallback((msg) => {
    stopJob()
    setStatus(msg)
    setError(true)
    setLoading(false)
  }, [stopJob])

  const pickFile = useCallback((f) => {
    if (!f) return
    setFile(f)
    clearState()
  }, [clearState])

  const cancel = useCallback(() => {
    stopJob()
    setLoading(false)
    clearState()
  }, [stopJob, clearState])

  const reset = useCallback(() => {
    stopJob()
    setFile(null)
    setLoading(false)
    clearState()
  }, [stopJob, clearState])

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
      jobIdRef.current = job_id

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
            jobIdRef.current = null
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
    pickFile, submit, reset, cancel,
    canReset: !!(file || status || downloadUrl),
  }
}
