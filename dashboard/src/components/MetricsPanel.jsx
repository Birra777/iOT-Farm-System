const LABEL = {
  'soil.moisture':       'Moisture',
  'soil.ph':             'Soil pH',
  'soil.nitrogen':       'Nitrogen',
  'weather.temperature': 'Temperature',
  'weather.humidity':    'Humidity',
  'weather.rainfall':    'Rainfall',
  'weather.wind_speed':  'Wind Speed',
}

export default function MetricsPanel({ summary }) {
  if (!summary) return <p className="loading">Select a field…</p>
  const { field, readings } = summary

  const soil    = readings?.filter(r => r.sensor_type === 'soil')    ?? []
  const weather = readings?.filter(r => r.sensor_type === 'weather') ?? []

  // Deduplicate: keep latest per metric
  const latest = {}
  for (const r of readings ?? []) {
    if (!latest[r.metric] || new Date(r.timestamp) > new Date(latest[r.metric].timestamp)) {
      latest[r.metric] = r
    }
  }
  const rows = Object.values(latest)

  if (rows.length === 0) return <p className="empty">No readings yet for {field?.name}</p>

  return (
    <div>
      {rows.map(r => (
        <div className="metric-row" key={r.metric}>
          <span className="metric-name">{LABEL[r.metric] ?? r.metric}</span>
          <span>
            <span className="metric-value">{r.value}</span>
            <span className="metric-unit">{r.unit}</span>
          </span>
        </div>
      ))}
    </div>
  )
}
