export default function FieldSelector({ fields, selected, onSelect }) {
  if (!fields) return <p className="loading">Loading fields…</p>

  return (
    <div>
      {fields.map(f => (
        <button
          key={f.id}
          className={`field-btn ${selected?.id === f.id ? 'active' : ''}`}
          onClick={() => onSelect(f)}
        >
          <strong>{f.name}</strong>
          <span className="crop">{f.crop_type}</span>
          <span className="ha">{f.hectares} ha · {f.zone_code}</span>
        </button>
      ))}
    </div>
  )
}
