export default function ProgressCard({ percent }) {
  return (
    <div className="loading-card" role="status" aria-live="polite">
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
      <p className="loading-title">Processing… {percent}%</p>
    </div>
  )
}
