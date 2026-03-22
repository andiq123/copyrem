import { XCircle } from 'lucide-react'

export default function ProgressCard({ percent, onCancel }) {
  return (
    <div style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: '0.5rem' }} role="status" aria-live="polite">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <p className="intensity-title" style={{ margin: 0 }}>
          Processing… <span style={{ color: 'var(--accent)', fontFamily: 'var(--font-mono)', marginLeft: '0.5rem' }}>{percent}%</span>
        </p>
        <button 
          onClick={onCancel}
          style={{ background: 'none', border: 'none', color: 'var(--text-dim)', cursor: 'pointer', display: 'flex', padding: 0 }}
          aria-label="Abort Task"
          className="hover-opacity"
        >
          <XCircle size={16} />
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

