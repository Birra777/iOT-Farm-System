const LABEL = {
  'soil.moisture':       'Moisture',
  'weather.temperature': 'Temp',
  'soil.nitrogen':       'Nitrogen',
  'weather.humidity':    'Humidity',
}
const PRIORITY = ['soil.moisture', 'weather.temperature', 'soil.nitrogen', 'weather.humidity']

export default function AllFieldsSnapshot({ fields, summaries }) {
  if (!fields || fields.length === 0) return <p className="loading">Loading…</p>

  return (
    <div className="snapshot-grid">
      {fields.map(f => {
        const s = summaries[f.id]
        const latest = {}
        for (const r of s?.readings ?? []) {
          if (!latest[r.metric]) latest[r.metric] = r
        }
        return (
          <div className="snapshot-cell" key={f.id}>
            <div className="field-name">{f.name}</div>
            {PRIORITY.map(m => latest[m] ? (
              <div className="snap-metric" key={m}>
                <span>{LABEL[m]}</span>
                <span className="snap-val">{latest[m].value} {latest[m].unit}</span>
              </div>
            ) : null)}
          </div>
        )
      })}
    </div>
  )
}
