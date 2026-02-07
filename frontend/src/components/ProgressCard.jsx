export default function ProgressCard({ percent }) {
  return (
    <div className="loading-card" role="status" aria-live="polite">
      <div className="progress-bar">
        <div className="progress-fill" style={{ width: `${percent}%` }} />
      </div>
      <p className="loading-title">Processing&hellip; {percent}%</p>
    </div>
  )
}
