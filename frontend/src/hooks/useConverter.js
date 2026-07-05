import { useState, useCallback, useEffect, useRef } from 'react'
import { useWebHaptics } from 'web-haptics/react'

const ACCEPT = '.mp3,.m4a,.wav,.flac,.aac,.ogg'
const SUFFIX = '_modified.mp3'
export const MAX_UPLOAD_MB = 80

export default function useConverter() {
  const haptic = useWebHaptics()
  const [file, setFile] = useState(null)
  const [loading, setLoading] = useState(false)
  const [percent, setPercent] = useState(0)
  const [status, setStatus] = useState(null)
  const [error, setError] = useState(false)
  const [jobId, setJobId] = useState(null)
  const [ready, setReady] = useState(false)
  const [downloadName, setDownloadName] = useState(null)
  const esRef = useRef(null)
  const jobIdRef = useRef(null)

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
    setReady(false)
    setDownloadName(null)
  }, [])

  const stopJob = useCallback(() => {
    if (jobIdRef.current) {
      fetch(`/convert/cancel/${jobIdRef.current}`, { method: 'POST' }).catch(() => {})
    }
    closeES()
    jobIdRef.current = null
    setJobId(null)
  }, [closeES])

  const fail = useCallback((msg) => {
    haptic.trigger('error', { intensity: 0.9 })
    stopJob()
    setStatus(msg)
    setError(true)
    setLoading(false)
    setReady(false)
  }, [stopJob, haptic])

  const pickFile = useCallback((f) => {
    if (!f) return
    setFile(f)
    stopJob()
    clearState()
  }, [clearState, stopJob])

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

  const submit = useCallback(async (intensity = 1.0) => {
    if (!file) return

    closeES()
    stopJob()
    setLoading(true)
    clearState()

    const form = new FormData()
    form.append('file', file)
    form.append('intensity', intensity.toString())

    try {
      const res = await fetch('/convert', { method: 'POST', body: form })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || res.statusText || 'Conversion failed')
      }
      const { job_id } = await res.json()
      jobIdRef.current = job_id
      setJobId(job_id)

      const es = new EventSource(`/convert/progress/${job_id}`)
      esRef.current = es

      es.onmessage = (event) => {
        const msg = JSON.parse(event.data)
        setPercent(msg.percent || 0)

        if (msg.error) return fail(msg.error)

        if (msg.done) {
          closeES()
          setPercent(100)
          setDownloadName(msg.filename?.trim() || `audio${SUFFIX}`)
          setStatus('Ready — preview your track below.')
          setReady(true)
          setLoading(false)
          jobIdRef.current = null
          haptic.trigger('success')
          setTimeout(() => haptic.trigger('heavy'), 120)
        }
      }

      es.onerror = () => fail('Connection lost. Please try again.')
    } catch (err) {
      fail(err.message || 'Something went wrong. Try again.')
    }
  }, [file, closeES, clearState, fail, haptic, stopJob])

  const previewOriginal = jobId ? `/convert/preview/${jobId}/original` : null
  const previewRemixed = ready && jobId ? `/convert/preview/${jobId}/remixed` : null
  const downloadHref = ready && jobId ? `/convert/download/${jobId}` : null

  return {
    file, loading, percent, status, error, ready,
    jobId, previewOriginal, previewRemixed, downloadHref, downloadName,
    accept: ACCEPT,
    pickFile, submit, reset, cancel,
    canReset: !!(file || status || ready),
  }
}
