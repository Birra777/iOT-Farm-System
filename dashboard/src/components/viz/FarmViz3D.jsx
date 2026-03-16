import { Suspense } from 'react'
import { Canvas } from '@react-three/fiber'
import { OrbitControls } from '@react-three/drei'
import SceneExtras from './SceneExtras'
import FieldPlot   from './FieldPlot'

// Field positions in 3D space.
// Coordinates are (x, z) centre points; width/depth in Three.js units (~1.5 ha each).
// Dry Ridge is slightly elevated (y=0.4) to represent terrain height.
const FIELD_LAYOUT = {
  NB: { x: -4.5, z: -3.5, w: 6.5, d: 5.2, elevation: 0   },
  RB: { x:  3.5, z: -2.5, w: 5.3, d: 4.5, elevation: 0   },
  DR: { x: -3.5, z:  3.0, w: 4.4, d: 3.6, elevation: 0.4 },
  SP: { x:  3.2, z:  3.2, w: 5.8, d: 4.8, elevation: 0   },
}

// Extract the latest value per metric from a summary's readings array.
function getLatestReadings(summary) {
  const latest = {}
  for (const r of summary?.readings ?? []) {
    if (!latest[r.metric] || new Date(r.timestamp) > new Date(latest[r.metric].timestamp)) {
      latest[r.metric] = r
    }
  }
  // Return plain metric → value map
  const out = {}
  for (const [metric, r] of Object.entries(latest)) {
    out[metric] = parseFloat(r.value)
  }
  return out
}

export default function FarmViz3D({ fields, summaries, alerts, selectedField, onSelectField }) {
  if (!fields || fields.length === 0) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center',
                    height: '100%', color: '#8b949e', fontSize: 13 }}>
        Loading field data…
      </div>
    )
  }

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      {/* Legend */}
      <div style={{
        position: 'absolute', top: 10, right: 10, zIndex: 10,
        background: '#1c2230cc', border: '1px solid #2a3441',
        borderRadius: 6, padding: '8px 12px', fontSize: 11, color: '#8b949e',
        backdropFilter: 'blur(4px)',
      }}>
        <div style={{ fontWeight: 600, marginBottom: 6, color: '#e6edf3', fontSize: 10, textTransform: 'uppercase', letterSpacing: '0.8px' }}>Soil Moisture</div>
        {[
          { color: '#b71c1c', label: '< 20%  Critical' },
          { color: '#ef6c00', label: '20–30% Warning' },
          { color: '#f9a825', label: '30–50% Dry' },
          { color: '#388e3c', label: '50–70% Optimal' },
          { color: '#1565c0', label: '> 70%  Wet' },
        ].map(e => (
          <div key={e.label} style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 3 }}>
            <div style={{ width: 10, height: 10, borderRadius: 2, background: e.color, flexShrink: 0 }} />
            {e.label}
          </div>
        ))}
        <div style={{ marginTop: 8, paddingTop: 6, borderTop: '1px solid #2a3441', fontWeight: 600, color: '#e6edf3', fontSize: 10, textTransform: 'uppercase', letterSpacing: '0.8px' }}>Ring = Nitrogen</div>
        {[
          { color: '#3fb950', label: 'Good  ≥100 mg/kg' },
          { color: '#d29922', label: 'Low  80–100' },
          { color: '#f85149', label: 'Deficient < 80' },
        ].map(e => (
          <div key={e.label} style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 3 }}>
            <div style={{ width: 10, height: 10, borderRadius: 2, background: e.color, flexShrink: 0 }} />
            {e.label}
          </div>
        ))}
        <div style={{ marginTop: 6, color: '#e6edf3', fontSize: 10 }}>
          💡 Hover field for readings<br />
          🖱 Click to select field<br />
          🔄 Drag to orbit · Scroll to zoom
        </div>
      </div>

      <Canvas
        camera={{ position: [1.5, 16, 21], fov: 38 }}
        style={{ width: '100%', height: '100%' }}
        shadows
        gl={{ antialias: true, toneMapping: 4, toneMappingExposure: 1.15 }}
      >
        <Suspense fallback={null}>
          <SceneExtras />

          <OrbitControls
            maxPolarAngle={Math.PI / 2.2}
            minDistance={4}
            maxDistance={35}
            target={[0, 0, 0]}
            enableDamping
            dampingFactor={0.08}
          />

          {fields.map(f => {
            const layout = FIELD_LAYOUT[f.zone_code]
            if (!layout) return null
            const readings    = getLatestReadings(summaries[f.id])
            const fieldAlerts = (alerts ?? []).filter(a => a.field_id === f.id)
            return (
              <FieldPlot
                key={f.id}
                field={f}
                layout={layout}
                readings={readings}
                alerts={fieldAlerts}
                selected={selectedField?.id === f.id}
                onSelectField={onSelectField}
              />
            )
          })}
        </Suspense>
      </Canvas>
    </div>
  )
}
