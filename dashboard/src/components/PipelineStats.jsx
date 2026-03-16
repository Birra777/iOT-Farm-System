export default function PipelineStats({ stats }) {
  if (!stats) return <p className="loading">Loading stats…</p>

  return (
    <div className="stat-grid">
      <div className="stat-cell">
        <div className="stat-value">{stats.total_readings_archived?.toLocaleString()}</div>
        <div className="stat-label">Total Archived</div>
      </div>
      <div className="stat-cell">
        <div className="stat-value">{stats.readings_last_minute}</div>
        <div className="stat-label">Readings/min</div>
      </div>
      <div className="stat-cell">
        <div className="stat-value">{stats.readings_last_hour?.toLocaleString()}</div>
        <div className="stat-label">Last Hour</div>
      </div>
      <div className="stat-cell">
        <div className="stat-value" style={{ color: stats.active_alerts > 0 ? 'var(--red)' : 'var(--green)' }}>
          {stats.active_alerts}
        </div>
        <div className="stat-label">Active Alerts</div>
      </div>
    </div>
  )
}
