import { XCircle } from 'lucide-react'

export default function ProgressCard({ percent, onCancel }) {
  return (
    <div className="progress-card" role="status" aria-live="polite">
      <div className="progress-header">
        <p className="progress-label">
          Processing… <span className="progress-pct">{percent}%</span>
        </p>
        <button type="button" className="btn-icon" onClick={onCancel} aria-label="Cancel">
          <XCircle size={18} aria-hidden="true" />
        </button>
      </div>
      <div
        className="progress-bar"
        role="progressbar"
        aria-valuenow={percent}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-label="Conversion progress"
      >
        <div className="progress-fill" style={{ width: `${percent}%` }} />
      </div>
    </div>
  )
}
