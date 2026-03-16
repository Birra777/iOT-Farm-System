import { resolveAlert } from '../api'

export default function AlertsPanel({ alerts, onRefresh }) {
  if (!alerts) return <p className="loading">Loading alerts…</p>
  if (alerts.length === 0) return <p className="empty">No active alerts</p>

  const handleResolve = async (id) => {
    try {
      await resolveAlert(id)
      onRefresh()
    } catch (e) {
      console.error(e)
    }
  }

  return (
    <div>
      {alerts.slice(0, 20).map(a => (
        <div key={a.id} className={`alert-item ${a.severity}`}>
          <div className="alert-header">
            <span className="alert-metric">{a.metric}</span>
            <span className={`badge ${a.severity}`}>{a.severity.toUpperCase()}</span>
          </div>
          <div className="alert-msg">{a.message}</div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ fontSize: 10, color: 'var(--text-secondary)' }}>
              {new Date(a.triggered_at).toLocaleTimeString()}
            </span>
            <button className="resolve-btn" onClick={() => handleResolve(a.id)}>
              Resolve
            </button>
          </div>
        </div>
      ))}
    </div>
  )
}
