import { useState, useEffect, useCallback, useRef } from 'react'
import { fetchNotifications, markNotificationRead, markAllNotificationsRead } from '../api'

const POLL_MS = 10_000

function timeAgo(iso) {
  const secs = Math.floor((Date.now() - new Date(iso)) / 1000)
  if (secs < 60)  return `${secs}s ago`
  if (secs < 3600) return `${Math.floor(secs / 60)}m ago`
  if (secs < 86400) return `${Math.floor(secs / 3600)}h ago`
  return `${Math.floor(secs / 86400)}d ago`
}

export default function NotificationBell() {
  const [data, setData]       = useState({ notifications: [], unread_count: 0 })
  const [open, setOpen]       = useState(false)
  const dropdownRef           = useRef(null)

  const load = useCallback(async () => {
    try {
      const res = await fetchNotifications({ limit: 30 })
      setData(res)
    } catch { /* silently ignore — API may not be up yet */ }
  }, [])

  useEffect(() => {
    load()
    const id = setInterval(load, POLL_MS)
    return () => clearInterval(id)
  }, [load])

  // Close when clicking outside
  useEffect(() => {
    if (!open) return
    function handler(e) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  async function handleRead(id) {
    await markNotificationRead(id)
    setData(prev => ({
      ...prev,
      unread_count: Math.max(0, prev.unread_count - 1),
      notifications: prev.notifications.map(n => n.id === id ? { ...n, is_read: true } : n),
    }))
  }

  async function handleReadAll() {
    await markAllNotificationsRead()
    setData(prev => ({
      ...prev,
      unread_count: 0,
      notifications: prev.notifications.map(n => ({ ...n, is_read: true })),
    }))
  }

  const { notifications, unread_count } = data
  const hasUnread = unread_count > 0

  return (
    <div style={{ position: 'relative' }} ref={dropdownRef}>
      <button
        onClick={() => setOpen(v => !v)}
        style={{
          background: 'none',
          border: '1px solid var(--border)',
          borderRadius: 5,
          color: hasUnread ? 'var(--yellow)' : 'var(--text-secondary)',
          padding: '4px 10px',
          cursor: 'pointer',
          fontSize: 14,
          display: 'flex',
          alignItems: 'center',
          gap: 5,
          transition: 'all 0.15s',
          borderColor: hasUnread ? 'var(--yellow)' : 'var(--border)',
          position: 'relative',
        }}
        title="Notifications"
      >
        🔔
        {hasUnread && (
          <span style={{
            background: 'var(--red)',
            color: '#fff',
            borderRadius: 9,
            fontSize: 10,
            fontWeight: 700,
            padding: '1px 5px',
            lineHeight: 1.4,
          }}>
            {unread_count > 99 ? '99+' : unread_count}
          </span>
        )}
      </button>

      {open && (
        <div style={{
          position: 'absolute',
          top: 'calc(100% + 6px)',
          right: 0,
          width: 360,
          maxHeight: 480,
          overflowY: 'auto',
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 8,
          boxShadow: '0 8px 32px rgba(0,0,0,0.45)',
          zIndex: 9999,
        }}>
          {/* Header */}
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            padding: '10px 14px',
            borderBottom: '1px solid var(--border)',
          }}>
            <span style={{ fontWeight: 600, fontSize: 13 }}>
              Notifications {hasUnread && <span style={{ color: 'var(--red)' }}>({unread_count} new)</span>}
            </span>
            {hasUnread && (
              <button onClick={handleReadAll} style={{
                background: 'none', border: 'none', color: 'var(--blue)',
                cursor: 'pointer', fontSize: 11, padding: 0,
              }}>
                Mark all read
              </button>
            )}
          </div>

          {/* List */}
          {notifications.length === 0 ? (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-secondary)', fontSize: 13 }}>
              No notifications yet
            </div>
          ) : (
            notifications.map(n => (
              <div
                key={n.id}
                onClick={() => !n.is_read && handleRead(n.id)}
                style={{
                  padding: '10px 14px',
                  borderBottom: '1px solid var(--border)',
                  cursor: n.is_read ? 'default' : 'pointer',
                  background: n.is_read ? 'transparent' : 'rgba(255,255,255,0.03)',
                  transition: 'background 0.1s',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8 }}>
                  <div style={{
                    fontWeight: n.is_read ? 400 : 600,
                    fontSize: 12,
                    color: n.severity === 'critical' ? 'var(--red)' : 'var(--yellow)',
                    lineHeight: 1.35,
                    flex: 1,
                  }}>
                    {n.title}
                  </div>
                  <div style={{ fontSize: 10, color: 'var(--text-secondary)', whiteSpace: 'nowrap', marginTop: 1 }}>
                    {timeAgo(n.created_at)}
                    {!n.is_read && (
                      <span style={{
                        display: 'inline-block', width: 6, height: 6,
                        borderRadius: '50%', background: 'var(--blue)',
                        marginLeft: 5, verticalAlign: 'middle',
                      }} />
                    )}
                  </div>
                </div>
                <div style={{ fontSize: 11, color: 'var(--text-secondary)', marginTop: 4, lineHeight: 1.45 }}>
                  {n.body}
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}
