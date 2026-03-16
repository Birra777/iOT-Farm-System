import { useState, useCallback } from 'react'
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'
import { usePolling } from '../hooks/usePolling'
import { fetchHistory } from '../api'

const METRICS = [
  { key: 'soil.moisture',       label: 'Moisture' },
  { key: 'soil.ph',             label: 'pH' },
  { key: 'soil.nitrogen',       label: 'Nitrogen' },
  { key: 'weather.temperature', label: 'Temp' },
  { key: 'weather.humidity',    label: 'Humidity' },
]

function fmt(ts) {
  return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export default function HistoryChart({ fieldId }) {
  const [metric, setMetric] = useState('soil.moisture')

  const fetcher = useCallback(
    () => fieldId ? fetchHistory(fieldId, metric) : Promise.resolve(null),
    [fieldId, metric]
  )
  const { data } = usePolling(fetcher, 10000)

  const points = (data?.readings ?? [])
    .map(r => ({ t: fmt(r.timestamp), v: parseFloat(r.value) }))
    .reverse()

  return (
    <div>
      <div className="metric-tabs">
        {METRICS.map(m => (
          <button
            key={m.key}
            className={`metric-tab ${metric === m.key ? 'active' : ''}`}
            onClick={() => setMetric(m.key)}
          >
            {m.label}
          </button>
        ))}
      </div>
      {points.length === 0 ? (
        <p className="empty">No history data yet</p>
      ) : (
        <ResponsiveContainer width="100%" height={180}>
          <LineChart data={points} margin={{ top: 4, right: 8, bottom: 0, left: -20 }}>
            <CartesianGrid stroke="#2a3441" strokeDasharray="3 3" />
            <XAxis dataKey="t" tick={{ fontSize: 10, fill: '#8b949e' }} interval="preserveStartEnd" />
            <YAxis tick={{ fontSize: 10, fill: '#8b949e' }} />
            <Tooltip
              contentStyle={{ background: '#1c2230', border: '1px solid #2a3441', fontSize: 11 }}
              labelStyle={{ color: '#8b949e' }}
            />
            <Line type="monotone" dataKey="v" stroke="#58a6ff" dot={false} strokeWidth={2} />
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  )
}
