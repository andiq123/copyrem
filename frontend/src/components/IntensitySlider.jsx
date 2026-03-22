import { useWebHaptics } from 'web-haptics/react'

export default function IntensitySlider({ value, onChange, disabled }) {
  const haptic = useWebHaptics()

  return (
    <div 
      className="intensity-section"
      style={{ 
        opacity: disabled ? 0.3 : 1, 
        pointerEvents: disabled ? 'none' : 'auto', 
        transition: 'opacity 0.4s var(--ease-out-expo)' 
      }}
    >
      <div className="intensity-header">
        <label htmlFor="intensity-slider" className="intensity-title">Engine Intensity</label>
        <span className="intensity-display">{value.toFixed(2)}x</span>
      </div>
      <input
        id="intensity-slider"
        type="range"
        min="0.5"
        max="2.5"
        step="0.05"
        value={value}
        onChange={(e) => {
          const val = parseFloat(e.target.value)
          onChange(val)
          // Feedback at neutral (1.0) and extremes
          if (Math.abs(val - 1.0) < 0.01) haptic.trigger('soft')
          if (val > 2.45) haptic.trigger('heavy')
        }}
        disabled={disabled}
        className="custom-slider"
      />
      <div className="intensity-indicators">
        <span>Subtle</span>
        <span>Standard</span>
        <span>Maximum</span>
      </div>
    </div>
  )
}
