import { AlertTriangle, CheckCircle2 } from 'lucide-react'

export default function StatusMessage({ message, isError }) {
  if (!message) return null

  return (
    <div className={`status-block ${isError ? 'is-error' : 'is-success'}`} role="alert" aria-live="assertive">
      <span className="status-icon" style={{ display: 'flex', alignItems: 'center' }}>
        {isError ? <AlertTriangle size={18} /> : <CheckCircle2 size={18} />}
      </span>
      <span>{message}</span>
    </div>
  )
}
