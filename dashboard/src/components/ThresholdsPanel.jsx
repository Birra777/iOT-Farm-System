import { useState, useEffect, useCallback } from 'react'
import { fetchThresholds, updateThreshold } from '../api'

const METRIC_LABELS = {
  'soil.moisture':        'Soil Moisture (%)',
  'soil.ph':              'Soil pH',
  'soil.nitrogen':        'Soil Nitrogen (mg/kg)',
  'weather.temperature':  'Air Temperature (°C)',
  'weather.humidity':     'Humidity (%)',
}

function numOrNull(v) {
  const n = parseFloat(v)
  return isNaN(n) ? null : n
}

export default function ThresholdsPanel({ onClose }) {
  const [thresholds, setThresholds] = useState([])
  const [saving, setSaving] = useState({})
  const [saved, setSaved] = useState({})
  const [edits, setEdits] = useState({})

  const load = useCallback(async () => {
    try {
      const data = await fetchThresholds()
      setThresholds(data)
      const init = {}
      for (const t of data) {
        init[t.metric] = {
          warning_low:   t.warning_low  ?? '',
          critical_low:  t.critical_low ?? '',
          warning_high:  t.warning_high ?? '',
          critical_high: t.critical_high ?? '',
        }
      }
      setEdits(init)
    } catch (e) {
      console.error('Failed to load thresholds', e)
    }
  }, [])

  useEffect(() => { load() }, [load])

  function handleChange(metric, field, value) {
    setEdits(prev => ({
      ...prev,
      [metric]: { ...prev[metric], [field]: value },
    }))
  }

  async function handleSave(metric) {
    setSaving(prev => ({ ...prev, [metric]: true }))
    try {
      const e = edits[metric] || {}
      await updateThreshold(metric, {
        warning_low:   numOrNull(e.warning_low),
        critical_low:  numOrNull(e.critical_low),
        warning_high:  numOrNull(e.warning_high),
        critical_high: numOrNull(e.critical_high),
      })
      setSaved(prev => ({ ...prev, [metric]: true }))
      setTimeout(() => setSaved(prev => ({ ...prev, [metric]: false })), 2000)
    } catch (e) {
      console.error('Save failed', e)
    } finally {
      setSaving(prev => ({ ...prev, [metric]: false }))
    }
  }

  return (
    <div className="threshold-overlay" onClick={onClose}>
      <div className="threshold-panel" onClick={e => e.stopPropagation()}>
        <div className="threshold-header">
          <span>⚙ Alert Thresholds</span>
          <button className="threshold-close" onClick={onClose}>✕</button>
        </div>
        <p className="threshold-desc">
          Set global warning and critical thresholds. Leave blank to disable a bound.
          Changes take effect within 5 minutes.
        </p>
        {thresholds.map(t => {
          const e = edits[t.metric] || {}
          return (
            <div key={t.metric} className="threshold-row">
              <div className="threshold-metric">{METRIC_LABELS[t.metric] || t.metric}</div>
              <div className="threshold-fields">
                {[
                  ['warning_low',   'Warn Low'],
                  ['critical_low',  'Crit Low'],
                  ['warning_high',  'Warn High'],
                  ['critical_high', 'Crit High'],
                ].map(([field, label]) => (
                  <label key={field} className="threshold-input-group">
                    <span>{label}</span>
                    <input
                      type="number"
                      step="any"
                      value={e[field] ?? ''}
                      placeholder="—"
                      onChange={ev => handleChange(t.metric, field, ev.target.value)}
                    />
                  </label>
                ))}
              </div>
              <button
                className={`threshold-save${saved[t.metric] ? ' saved' : ''}`}
                onClick={() => handleSave(t.metric)}
                disabled={saving[t.metric]}
              >
                {saved[t.metric] ? 'Saved ✓' : saving[t.metric] ? '…' : 'Save'}
              </button>
            </div>
          )
        })}
      </div>
    </div>
  )
}
