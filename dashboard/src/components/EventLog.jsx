export default function EventLog({ alerts }) {
  const events = (alerts ?? []).slice(0, 30)

  if (events.length === 0) return <p className="empty">No events yet</p>

  return (
    <div className="event-log">
      {events.map(a => (
        <div className="event-entry" key={a.id}>
          <span className="evt-time">
            {new Date(a.triggered_at).toLocaleTimeString()}
          </span>
          <span className={`evt-sev-${a.severity}`}>[{a.severity.toUpperCase()}]</span>
          {' '}{a.metric} → {a.value}
        </div>
      ))}
    </div>
  )
}
