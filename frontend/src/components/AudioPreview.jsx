import { useRef, useState, useEffect, useCallback } from 'react'
import { Play, Pause } from 'lucide-react'
import { formatTime } from '../utils/formatSize'

export default function AudioPreview({ src, title, badge, active }) {
  const audioRef = useRef(null)
  const [playing, setPlaying] = useState(false)
  const [time, setTime] = useState(0)
  const [duration, setDuration] = useState(0)

  useEffect(() => {
    setPlaying(false)
    setTime(0)
    setDuration(0)
  }, [src])

  const toggle = useCallback(() => {
    const a = audioRef.current
    if (!a) return
    if (a.paused) a.play()
    else a.pause()
  }, [])

  const seek = useCallback((e) => {
    const a = audioRef.current
    if (!a || !duration) return
    const { left, width } = e.currentTarget.getBoundingClientRect()
    a.currentTime = Math.max(0, Math.min(1, (e.clientX - left) / width)) * duration
  }, [duration])

  const pct = duration ? (time / duration) * 100 : 0

  return (
    <div className={`player${active ? ' player-active' : ''}${playing ? ' player-playing' : ''}`}>
      <audio
        ref={audioRef}
        src={src}
        preload="metadata"
        onPlay={() => setPlaying(true)}
        onPause={() => setPlaying(false)}
        onEnded={() => setPlaying(false)}
        onLoadedMetadata={(e) => setDuration(e.currentTarget.duration)}
        onTimeUpdate={(e) => setTime(e.currentTarget.currentTime)}
      />

      <button type="button" className="player-btn" onClick={toggle} aria-label={playing ? 'Pause' : 'Play'}>
        {playing ? <Pause size={18} fill="currentColor" /> : <Play size={18} fill="currentColor" />}
      </button>

      <div className="player-body">
        <div className="player-meta">
          {badge && <span className="player-badge">{badge}</span>}
          <span className="player-title">{title}</span>
        </div>

        <div
          className="player-scrub"
          role="slider"
          aria-valuemin={0}
          aria-valuemax={duration}
          aria-valuenow={time}
          aria-label="Seek"
          onClick={seek}
        >
          <div className="player-scrub-fill" style={{ width: `${pct}%` }} />
        </div>

        <div className="player-times">
          <span>{formatTime(time)}</span>
          <span>{formatTime(duration)}</span>
        </div>
      </div>

      <div className="player-bars" aria-hidden="true">
        {[0, 1, 2, 3, 4].map((i) => (
          <span key={i} className="player-bar" style={{ animationDelay: `${i * 0.12}s` }} />
        ))}
      </div>
    </div>
  )
}
