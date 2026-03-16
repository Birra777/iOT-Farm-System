import { useState, useCallback, useEffect } from 'react'
import { usePolling } from './hooks/usePolling'
import { useSSE } from './hooks/useSSE'
import { fetchFields, fetchSummary, fetchAlerts, fetchStats } from './api'
import FieldSelector     from './components/FieldSelector'
import ServiceHealth     from './components/ServiceHealth'
import MetricsPanel      from './components/MetricsPanel'
import HistoryChart      from './components/HistoryChart'
import AllFieldsSnapshot from './components/AllFieldsSnapshot'
import AlertsPanel       from './components/AlertsPanel'
import PipelineStats     from './components/PipelineStats'
import EventLog          from './components/EventLog'
import FarmViz3D         from './components/viz/FarmViz3D'
import NotificationBell  from './components/NotificationBell'
import ThresholdsPanel   from './components/ThresholdsPanel'
import FarmAdvisor       from './components/FarmAdvisor'

export default function App() {
  const [selectedField, setSelectedField] = useState(null)
  const [summaries, setSummaries] = useState({})
  const [show3D, setShow3D] = useState(false)
  const [showThresholds, setShowThresholds] = useState(false)

  // Core data feeds — polling kept as backup sync (SSE is primary for readings)
  const { data: fields } = usePolling(useCallback(() => fetchFields(), []), 30000)
  const { data: alerts, refresh: refreshAlerts } = usePolling(
    useCallback(() => fetchAlerts('active'), []), 30000
  )
  const { data: allAlerts } = usePolling(
    useCallback(() => fetchAlerts(''), []), 30000
  )
  const { data: stats } = usePolling(useCallback(() => fetchStats(), []), 15000)

  // Auto-select first field once loaded
  useEffect(() => {
    if (fields?.length && !selectedField) setSelectedField(fields[0])
  }, [fields, selectedField])

  // Selected field summary (polled — SSE updates summaries map but MetricsPanel uses this)
  const summaryFetcher = useCallback(
    () => selectedField ? fetchSummary(selectedField.id) : Promise.resolve(null),
    [selectedField]
  )
  const { data: summary, refresh: refreshSummary } = usePolling(summaryFetcher, 30000)

  // All-fields summaries — initial load then refreshed by SSE events
  useEffect(() => {
    if (!fields) return
    const load = async () => {
      const results = await Promise.allSettled(
        fields.map(f => fetchSummary(f.id).then(s => [f.id, s]))
      )
      const map = {}
      for (const r of results) {
        if (r.status === 'fulfilled') {
          const [id, s] = r.value
          map[id] = s
        }
      }
      setSummaries(map)
    }
    load()
    const id = setInterval(load, 30000)
    return () => clearInterval(id)
  }, [fields])

  // SSE — live readings push from server; refresh the affected field's summary
  useSSE(useCallback((eventType, data) => {
    if (eventType === 'reading' && data.field_id) {
      fetchSummary(data.field_id).then(s => {
        setSummaries(prev => ({ ...prev, [data.field_id]: s }))
        if (selectedField?.id === data.field_id) refreshSummary()
      }).catch(() => {})
    }
    if (eventType === 'alert') {
      refreshAlerts()
    }
  }, [selectedField, refreshSummary, refreshAlerts]))

  return (
    <>
      <header className="app-header">
        <h1>🌾 AgriStream</h1>
        <span className="subtitle">Kavango East · Real-time Farm Monitoring</span>
        <NotificationBell />
        <button
          className="threshold-gear"
          onClick={() => setShowThresholds(true)}
          title="Alert Thresholds"
        >⚙</button>
        <button
          className={`viz-toggle ${show3D ? 'active' : ''}`}
          onClick={() => setShow3D(v => !v)}
        >
          {show3D ? '◀ Dashboard' : '🗺 3D Farm'}
        </button>
      </header>
      {showThresholds && <ThresholdsPanel onClose={() => setShowThresholds(false)} />}

      <div className="panels">
        {/* LEFT — Field selector + health */}
        <div className="panel">
          <div className="card">
            <div className="card-title">Farm Fields</div>
            <FieldSelector
              fields={fields}
              selected={selectedField}
              onSelect={setSelectedField}
            />
          </div>
          <ServiceHealth />
        </div>

        {/* CENTRE — 3D view OR standard dashboard */}
        {show3D ? (
          <div className="panel viz-panel">
            <FarmViz3D
              fields={fields}
              summaries={summaries}
              alerts={alerts}
              selectedField={selectedField}
              onSelectField={setSelectedField}
            />
          </div>
        ) : (
          <div className="panel">
            <div className="card">
              <div className="card-title">
                {selectedField ? `${selectedField.name} · Latest Readings` : 'Latest Readings'}
              </div>
              <MetricsPanel summary={summary} />
            </div>

            <div className="card">
              <div className="card-title">
                {selectedField ? `${selectedField.name} · 30-min History` : '30-min History'}
              </div>
              <HistoryChart fieldId={selectedField?.id} />
            </div>

            <div className="card">
              <div className="card-title">All Fields Snapshot</div>
              <AllFieldsSnapshot fields={fields} summaries={summaries} />
            </div>
          </div>
        )}

        {/* RIGHT — Alerts + stats + log */}
        <div className="panel">
          <div className="card">
            <div className="card-title">Pipeline Stats</div>
            <PipelineStats stats={stats} />
          </div>

          <div className="card">
            <div className="card-title">AI Farm Advisor</div>
            <FarmAdvisor />
          </div>

          <div className="card">
            <div className="card-title">
              Active Alerts
              {alerts?.length ? <span style={{ color: 'var(--red)', marginLeft: 6 }}>({alerts.length})</span> : null}
            </div>
            <AlertsPanel alerts={alerts} onRefresh={refreshAlerts} />
          </div>

          <div className="card">
            <div className="card-title">Event Log</div>
            <EventLog alerts={allAlerts} />
          </div>
        </div>
      </div>
    </>
  )
}
