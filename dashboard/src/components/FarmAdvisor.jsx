import { useState, useRef } from 'react'
import { fetchAIAdvice } from '../api'

const CACHE_MS = 2 * 60 * 1000 // 2 minutes

export default function FarmAdvisor() {
  const [advice, setAdvice] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [fetchedAt, setFetchedAt] = useState(null)
  const cacheRef = useRef({ text: null, at: 0 })

  async function handleClick() {
    const now = Date.now()
    if (cacheRef.current.text && now - cacheRef.current.at < CACHE_MS) {
      setAdvice(cacheRef.current.text)
      setFetchedAt(new Date(cacheRef.current.at))
      return
    }

    setLoading(true)
    setError(null)
    try {
      const data = await fetchAIAdvice()
      cacheRef.current = { text: data.advice, at: Date.now() }
      setAdvice(data.advice)
      setFetchedAt(new Date())
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  const age = fetchedAt
    ? Math.round((Date.now() - fetchedAt.getTime()) / 1000)
    : null

  return (
    <div>
      <button className="advisor-btn" onClick={handleClick} disabled={loading}>
        {loading ? 'Consulting AI advisor…' : '🌿 Get AI Farm Advice'}
      </button>
      {error && <div className="advisor-error">{error}</div>}
      {advice && !error && (
        <>
          <div className="advisor-text">{advice}</div>
          {age !== null && (
            <div className="advisor-age">Generated {age}s ago</div>
          )}
        </>
      )}
    </div>
  )
}
