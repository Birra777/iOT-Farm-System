import { useEffect, useRef, useCallback } from 'react'

// useSSE subscribes to the server-sent events stream at /api/events.
// onEvent(eventType, data) is called for each named event received.
// Returns a cleanup function via useEffect.
export function useSSE(onEvent) {
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  const connect = useCallback(() => {
    const es = new EventSource('/api/events')

    es.addEventListener('reading', e => {
      try {
        onEventRef.current('reading', JSON.parse(e.data))
      } catch {}
    })

    es.addEventListener('alert', e => {
      try {
        onEventRef.current('alert', JSON.parse(e.data))
      } catch {}
    })

    es.onerror = () => {
      // EventSource auto-reconnects on error; nothing to do here.
    }

    return es
  }, [])

  useEffect(() => {
    const es = connect()
    return () => es.close()
  }, [connect])
}
