import { useWebHaptics } from 'web-haptics/react'

export default function IntensitySlider({ value, onChange, disabled }) {
  const haptic = useWebHaptics()

  return (
    <div className={`intensity-section${disabled ? ' is-disabled' : ''}`}>
      <div className="intensity-header">
        <label htmlFor="intensity-slider" className="intensity-title">Strength</label>
        <span className="intensity-display">{Math.round(value * 100)}%</span>
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
          if (Math.abs(val - 1.0) < 0.01) haptic.trigger('soft')
          if (val > 2.45) haptic.trigger('heavy')
        }}
        disabled={disabled}
        className="custom-slider"
      />
      <div className="intensity-indicators">
        <span>50%</span>
        <span>100%</span>
        <span>250%</span>
      </div>
    </div>
  )
}
