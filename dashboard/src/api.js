const BASE = ''

export async function fetchFields() {
  const r = await fetch(`${BASE}/api/fields`)
  if (!r.ok) throw new Error('fields fetch failed')
  return r.json()
}

export async function fetchSummary(fieldId) {
  const r = await fetch(`${BASE}/api/fields/${fieldId}/summary`)
  if (!r.ok) throw new Error('summary fetch failed')
  return r.json()
}

export async function fetchHistory(fieldId, metric) {
  const to = new Date()
  const from = new Date(to - 30 * 60 * 1000)
  const params = new URLSearchParams({
    metric,
    from: from.toISOString(),
    to: to.toISOString(),
  })
  const r = await fetch(`${BASE}/api/fields/${fieldId}/history?${params}`)
  if (!r.ok) throw new Error('history fetch failed')
  return r.json()
}

export async function fetchAlerts(status = 'active') {
  const r = await fetch(`${BASE}/api/alerts?status=${status}`)
  if (!r.ok) throw new Error('alerts fetch failed')
  return r.json()
}

export async function fetchStats() {
  const r = await fetch(`${BASE}/api/stats`)
  if (!r.ok) throw new Error('stats fetch failed')
  return r.json()
}

export async function fetchHealth() {
  try {
    const r = await fetch(`${BASE}/health`)
    return r.ok
  } catch {
    return false
  }
}

export async function resolveAlert(id) {
  const r = await fetch(`${BASE}/api/alerts/${id}/resolve`, { method: 'POST' })
  if (!r.ok) throw new Error('resolve failed')
  return r.json()
}

export async function fetchNotifications({ unread = false, limit = 50 } = {}) {
  const params = new URLSearchParams({ limit })
  if (unread) params.set('unread', 'true')
  const r = await fetch(`${BASE}/api/notifications?${params}`)
  if (!r.ok) throw new Error('notifications fetch failed')
  return r.json() // { notifications: [...], unread_count: N }
}

export async function markNotificationRead(id) {
  const r = await fetch(`${BASE}/api/notifications/${id}/read`, { method: 'POST' })
  if (!r.ok) throw new Error('mark read failed')
  return r.json()
}

export async function markAllNotificationsRead() {
  const r = await fetch(`${BASE}/api/notifications/read-all`, { method: 'POST' })
  if (!r.ok) throw new Error('mark all read failed')
  return r.json()
}

export async function fetchThresholds() {
  const r = await fetch(`${BASE}/api/thresholds`)
  if (!r.ok) throw new Error('thresholds fetch failed')
  return r.json()
}

export async function updateThreshold(metric, data) {
  const r = await fetch(`${BASE}/api/thresholds/${encodeURIComponent(metric)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!r.ok) throw new Error('threshold update failed')
  return r.json()
}

export async function fetchAIAdvice() {
  const r = await fetch(`${BASE}/api/advisor`, { method: 'POST' })
  if (!r.ok) throw new Error('advisor fetch failed')
  return r.json()
}
