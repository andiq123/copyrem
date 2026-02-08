import { useState, useCallback, useEffect, useRef } from 'react'

const DEFAULT_API_INFO = {
  max_upload_mb: 80,
  allowed_extensions: ['.mp3', '.m4a', '.wav', '.flac', '.aac', '.ogg'],
}

const VOCALS_SUFFIX = '_vocals.mp3'
const INSTRUMENTAL_SUFFIX = '_instrumental.mp3'

export default function useSeparate() {
  const [apiInfo, setApiInfo] = useState(DEFAULT_API_INFO)
  const [file, setFile] = useState(null)
  const [loading, setLoading] = useState(false)
  const [percent, setPercent] = useState(0)
  const [status, setStatus] = useState(null)
  const [error, setError] = useState(false)
  const [downloadVocalsUrl, setDownloadVocalsUrl] = useState(null)
  const [downloadVocalsName, setDownloadVocalsName] = useState(null)
  const [downloadInstrumentalUrl, setDownloadInstrumentalUrl] = useState(null)
  const [downloadInstrumentalName, setDownloadInstrumentalName] = useState(null)
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
        navigator.sendBeacon(`/separate/cancel/${jobIdRef.current}`)
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
    setDownloadVocalsUrl((u) => {
      if (u) URL.revokeObjectURL(u)
      return null
    })
    setDownloadVocalsName(null)
    setDownloadInstrumentalUrl((u) => {
      if (u) URL.revokeObjectURL(u)
      return null
    })
    setDownloadInstrumentalName(null)
  }, [])

  const stopJob = useCallback(() => {
    if (jobIdRef.current) {
      fetch(`/separate/cancel/${jobIdRef.current}`, { method: 'POST' }).catch(() => {})
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
      const res = await fetch('/separate', { method: 'POST', body: form })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || res.statusText || 'Separation failed')
      }
      const { job_id } = await res.json()
      jobIdRef.current = job_id

      const es = new EventSource(`/separate/progress/${job_id}`)
      esRef.current = es

      es.onmessage = async (event) => {
        const msg = JSON.parse(event.data)
        setPercent(msg.percent || 0)

        if (msg.error) return fail(msg.error)

        if (msg.done) {
          closeES()
          setPercent(100)
          try {
            const [dlVocals, dlInst] = await Promise.all([
              fetch(`/separate/download/${job_id}/vocals`),
              fetch(`/separate/download/${job_id}/instrumental`),
            ])
            if (!dlVocals.ok || !dlInst.ok) throw new Error('Download failed')
            const [blobVocals, blobInst] = await Promise.all([dlVocals.blob(), dlInst.blob()])
            const dispVocals = dlVocals.headers.get('Content-Disposition')
            const dispInst = dlInst.headers.get('Content-Disposition')
            const matchVocals = dispVocals?.match(/filename="?([^";]+)"?/)
            const matchInst = dispInst?.match(/filename="?([^";]+)"?/)
            setDownloadVocalsUrl(URL.createObjectURL(blobVocals))
            setDownloadVocalsName(matchVocals?.[1]?.trim() || `audio${VOCALS_SUFFIX}`)
            setDownloadInstrumentalUrl(URL.createObjectURL(blobInst))
            setDownloadInstrumentalName(matchInst?.[1]?.trim() || `audio${INSTRUMENTAL_SUFFIX}`)
            setStatus('Vocals and instrumental are ready.')
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
  }, [file, closeES, clearState, fail])

  const accept = (apiInfo?.allowed_extensions || DEFAULT_API_INFO.allowed_extensions).join(',')

  return {
    apiInfo, file, loading, percent, status, error,
    downloadVocalsUrl, downloadVocalsName,
    downloadInstrumentalUrl, downloadInstrumentalName,
    accept,
    pickFile, submit, reset, cancel,
    canReset: !!(file || status || downloadVocalsUrl || downloadInstrumentalUrl),
  }
}
