import { useState, useEffect } from 'react'
import { fetchHealth } from '../api'

export default function ServiceHealth() {
  const [ok, setOk] = useState(null)

  useEffect(() => {
    const check = async () => setOk(await fetchHealth())
    check()
    const id = setInterval(check, 10000)
    return () => clearInterval(id)
  }, [])

  return (
    <div className="card">
      <div className="card-title">Service Health</div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        {[
          { name: 'API Server', status: ok },
          { name: 'Simulator',  status: ok },
          { name: 'Processor',  status: ok },
          { name: 'Anomaly Det.', status: ok },
        ].map(s => (
          <div key={s.name} style={{ display: 'flex', alignItems: 'center', fontSize: 12 }}>
            <span className={`health-dot ${s.status === null ? '' : s.status ? 'ok' : 'err'}`} />
            {s.name}
          </div>
        ))}
      </div>
    </div>
  )
}
