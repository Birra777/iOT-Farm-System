import { useRef, useState, useMemo } from 'react'
import { useFrame } from '@react-three/fiber'
import { Html } from '@react-three/drei'
import * as THREE from 'three'
import RainParticles from './RainParticles'
import CropInstances from './CropInstances'

function moistureColor(v) {
  if (v == null) return '#4a5568'
  if (v < 20)   return '#b71c1c'
  if (v < 30)   return '#ef6c00'
  if (v < 50)   return '#f9a825'
  if (v < 70)   return '#388e3c'
  return         '#1565c0'
}
function nitrogenColor(v) {
  if (v == null) return '#4a5568'
  if (v < 80)   return '#f85149'
  if (v < 100)  return '#d29922'
  return         '#3fb950'
}

// ── Perimeter fence: wooden posts + 2 horizontal wire lines ──────────────────
function FieldFence({ w, d, elevation }) {
  const posts = useMemo(() => {
    const sp  = 0.9
    const pts = []
    for (let x = -w / 2; x <= w / 2 + 0.001; x += sp)
      pts.push([Math.min(x, w / 2), -d / 2], [Math.min(x, w / 2), d / 2])
    for (let z = -d / 2 + sp; z < d / 2 - 0.001; z += sp)
      pts.push([-w / 2, z], [w / 2, z])
    return pts
  }, [w, d])

  return (
    <group>
      {/* Posts */}
      {posts.map(([px, pz], i) => (
        <mesh key={i} position={[px, elevation + 0.24, pz]}>
          <cylinderGeometry args={[0.032, 0.036, 0.52, 5]} />
          <meshLambertMaterial color="#6d4c2a" />
        </mesh>
      ))}
      {/* Two wire lines per side */}
      {[0.10, 0.30].map((h, hi) => (
        <group key={hi}>
          <mesh position={[0,    elevation + h, -d / 2]}><boxGeometry args={[w, 0.01, 0.01]} /><meshLambertMaterial color="#8d6e63" /></mesh>
          <mesh position={[0,    elevation + h,  d / 2]}><boxGeometry args={[w, 0.01, 0.01]} /><meshLambertMaterial color="#8d6e63" /></mesh>
          <mesh position={[-w/2, elevation + h,  0]}><boxGeometry args={[0.01, 0.01, d]} /><meshLambertMaterial color="#8d6e63" /></mesh>
          <mesh position={[ w/2, elevation + h,  0]}><boxGeometry args={[0.01, 0.01, d]} /><meshLambertMaterial color="#8d6e63" /></mesh>
        </group>
      ))}
    </group>
  )
}

// ── Sparse trees outside fence ────────────────────────────────────────────────
function BorderTrees({ w, d, elevation }) {
  const trees = useMemo(() => {
    let s = 71
    const rand = () => { s = (s * 1664525 + 1013904223) & 0xffffffff; return (s >>> 0) / 0xffffffff }
    const pts = []
    for (let i = 0; i < 4; i++) {
      pts.push({ x: -w/2 + rand() * w,           z: -(d/2 + 0.45 + rand() * 0.5), h: 0.22 + rand() * 0.38, r: 0.10 + rand() * 0.08 })
      pts.push({ x: -w/2 + rand() * w,           z:   d/2 + 0.45 + rand() * 0.5,  h: 0.22 + rand() * 0.38, r: 0.10 + rand() * 0.08 })
      pts.push({ x: -(w/2 + 0.45 + rand() * 0.5), z: -d/2 + rand() * d,           h: 0.22 + rand() * 0.38, r: 0.10 + rand() * 0.08 })
      pts.push({ x:   w/2 + 0.45 + rand() * 0.5,  z: -d/2 + rand() * d,           h: 0.22 + rand() * 0.38, r: 0.10 + rand() * 0.08 })
    }
    return pts
  }, [w, d])

  return (
    <group>
      {trees.map((t, i) => (
        <group key={i} position={[t.x, elevation, t.z]}>
          {/* Trunk */}
          <mesh position={[0, t.h * 0.28, 0]}>
            <cylinderGeometry args={[0.025, 0.035, t.h * 0.55, 5]} />
            <meshLambertMaterial color="#5d4037" />
          </mesh>
          {/* Lower canopy — wider */}
          <mesh position={[0, t.h * 0.78, 0]}>
            <coneGeometry args={[t.r, t.h * 0.72, 6]} />
            <meshLambertMaterial color="#1b4a1b" />
          </mesh>
          {/* Upper canopy — narrower, lighter */}
          <mesh position={[0, t.h * 1.28, 0]}>
            <coneGeometry args={[t.r * 0.6, t.h * 0.5, 5]} />
            <meshLambertMaterial color="#245e24" />
          </mesh>
        </group>
      ))}
    </group>
  )
}

// ── Crop row furrows etched into top surface ──────────────────────────────────
function CropFurrows({ w, d, elevation, halfH, cropType }) {
  const color = { Maize: '#184010', Sorghum: '#3d2000', Millet: '#5a6808', Groundnuts: '#3a5808' }[cropType] ?? '#1a3a0e'
  const xcolor = { Maize: '#0e2a08', Sorghum: '#2a1400', Millet: '#3a4405', Groundnuts: '#263804' }[cropType] ?? '#112408'
  const rowSpacing = 0.32
  const rows = Math.floor((d - 0.28) / rowSpacing)
  const cols = Math.floor((w - 0.28) / (rowSpacing * 2.2))
  return (
    <group>
      {/* Plough rows along Z */}
      {Array.from({ length: rows }, (_, i) => (
        <mesh key={`r${i}`} position={[0, elevation + halfH * 2 + 0.007, -d/2 + 0.18 + i * rowSpacing]}>
          <boxGeometry args={[w - 0.16, 0.007, 0.038]} />
          <meshLambertMaterial color={color} transparent opacity={0.62} />
        </mesh>
      ))}
      {/* Cross-plough lines along X — sparser */}
      {Array.from({ length: cols }, (_, i) => (
        <mesh key={`c${i}`} position={[-w/2 + 0.18 + i * rowSpacing * 2.2, elevation + halfH * 2 + 0.005, 0]}>
          <boxGeometry args={[0.028, 0.005, d - 0.16]} />
          <meshLambertMaterial color={xcolor} transparent opacity={0.4} />
        </mesh>
      ))}
    </group>
  )
}

// ── Soil probe stakes (3 per field, in-ground sensors) ───────────────────────
function SoilProbes({ w, d, elevation, halfH }) {
  const positions = useMemo(() => {
    let s = 37
    const rand = () => { s = (s * 1664525 + 1013904223) & 0xffffffff; return (s >>> 0) / 0xffffffff }
    return Array.from({ length: 3 }, () => [(rand() - 0.5) * (w - 1.4), (rand() - 0.5) * (d - 1.4)])
  }, [w, d])

  const tipRefs = useRef([])
  useFrame(({ clock }) => {
    const t = clock.getElapsedTime()
    tipRefs.current.forEach((ref, i) => {
      if (ref) ref.material.emissiveIntensity = 0.75 + Math.sin(t * 1.5 + i * 2.1) * 0.3
    })
  })

  return (
    <group>
      {positions.map(([px, pz], i) => (
        <group key={i} position={[px, 0, pz]}>
          <mesh position={[0, elevation + halfH * 2 - 0.04, 0]}>
            <cylinderGeometry args={[0.018, 0.022, 0.22, 5]} />
            <meshLambertMaterial color="#6d4c2a" />
          </mesh>
          <mesh ref={el => { tipRefs.current[i] = el }} position={[0, elevation + halfH * 2 + 0.07, 0]}>
            <sphereGeometry args={[0.032, 7, 6]} />
            <meshStandardMaterial color="#ff6f00" emissive="#ff6f00" emissiveIntensity={0.75} />
          </mesh>
        </group>
      ))}
    </group>
  )
}

// ── Windmill (Sandveld Millet Ridge, DR field) ────────────────────────────────
function Windmill({ elevation, halfH, windSpeed, w }) {
  const bladeRef = useRef()
  const rpm      = Math.max(0.4, Math.min(2.8, (windSpeed ?? 10) / 12))

  useFrame((_, delta) => {
    if (bladeRef.current) bladeRef.current.rotation.z += rpm * delta
  })

  const base    = elevation + halfH * 2
  const towerH  = 3.4
  const towerX  = -(w / 2 + 1.8)

  return (
    <group position={[towerX, 0, 0]}>
      {/* 4 tapered legs */}
      {[[-0.3,-0.3],[0.3,-0.3],[-0.3,0.3],[0.3,0.3]].map(([lx,lz], i) => (
        <mesh key={i} position={[lx * 0.55, base + towerH / 2, lz * 0.55]}>
          <cylinderGeometry args={[0.022, 0.030, towerH, 4]} />
          <meshStandardMaterial color="#607d8b" roughness={0.28} metalness={0.82} />
        </mesh>
      ))}
      {/* Cross braces */}
      {[0.7, 1.5, 2.3].map((bh, i) => (
        <mesh key={i} position={[0, base + bh, 0]} rotation={[0, i * 0.7, 0]}>
          <boxGeometry args={[0.62, 0.016, 0.016]} />
          <meshStandardMaterial color="#78909c" roughness={0.3} metalness={0.8} />
        </mesh>
      ))}
      {/* Nacelle (housing) */}
      <mesh position={[0, base + towerH + 0.14, 0]}>
        <boxGeometry args={[0.2, 0.2, 0.38]} />
        <meshStandardMaterial color="#90a4ae" roughness={0.18} metalness={0.72} />
      </mesh>
      {/* Hub */}
      <mesh position={[0, base + towerH + 0.14, 0.22]}>
        <sphereGeometry args={[0.09, 10, 8]} />
        <meshStandardMaterial color="#cfd8dc" roughness={0.18} metalness={0.82} />
      </mesh>
      {/* Rotating blades */}
      <group ref={bladeRef} position={[0, base + towerH + 0.14, 0.3]}>
        {Array.from({ length: 6 }, (_, i) => {
          const a = (i / 6) * Math.PI * 2
          return (
            <mesh key={i} position={[Math.cos(a) * 0.44, Math.sin(a) * 0.44, 0]} rotation={[0, 0, a]}>
              <boxGeometry args={[0.055, 0.76, 0.016]} />
              <meshStandardMaterial color="#eceff1" roughness={0.22} />
            </mesh>
          )
        })}
      </group>
    </group>
  )
}

// ── IoT sensor tower: base + pole + T-bar + sensor pods + solar panel + anemometer + status orb ──
function SensorTower({ tx, tz, elevation, halfH, idx }) {
  const orbRef = useRef()
  const phase  = idx * 1.04

  useFrame(({ clock }) => {
    if (orbRef.current)
      orbRef.current.material.emissiveIntensity = 0.55 + Math.sin(clock.getElapsedTime() * 1.3 + phase) * 0.4
  })

  const base  = elevation + halfH * 2
  const poleH = 1.3

  return (
    <group position={[tx, 0, tz]}>
      {/* Base disk */}
      <mesh position={[0, base + 0.02, 0]}>
        <cylinderGeometry args={[0.1, 0.13, 0.04, 8]} />
        <meshStandardMaterial color="#374151" roughness={0.3} metalness={0.85} />
      </mesh>
      {/* Pole */}
      <mesh position={[0, base + poleH / 2 + 0.04, 0]}>
        <cylinderGeometry args={[0.021, 0.026, poleH, 6]} />
        <meshStandardMaterial color="#4b5563" roughness={0.2} metalness={0.92} />
      </mesh>
      {/* T crossbar */}
      <mesh position={[0, base + poleH * 0.76, 0]} rotation={[0, 0, Math.PI / 2]}>
        <cylinderGeometry args={[0.013, 0.013, 0.48, 5]} />
        <meshStandardMaterial color="#6b7280" roughness={0.2} metalness={0.92} />
      </mesh>
      {/* Soil sensor (left arm, blue box) */}
      <mesh position={[-0.22, base + poleH * 0.76, 0]}>
        <boxGeometry args={[0.068, 0.052, 0.034]} />
        <meshStandardMaterial color="#1d4ed8" roughness={0.12} metalness={0.5} emissive="#1e3a8a" emissiveIntensity={0.45} />
      </mesh>
      {/* Weather dome (right arm, teal sphere) */}
      <mesh position={[0.22, base + poleH * 0.76 + 0.02, 0]}>
        <sphereGeometry args={[0.054, 9, 7]} />
        <meshStandardMaterial color="#0f766e" roughness={0.08} metalness={0.5} emissive="#0d9488" emissiveIntensity={0.55} />
      </mesh>
      {/* Anemometer cups — 3 small white spheres on weather dome */}
      {[0, Math.PI * 2 / 3, Math.PI * 4 / 3].map((a, i) => (
        <mesh key={i} position={[0.22 + Math.cos(a) * 0.088, base + poleH * 0.76 + 0.085, Math.sin(a) * 0.088]}>
          <sphereGeometry args={[0.021, 6, 5]} />
          <meshLambertMaterial color="#f5f5f5" />
        </mesh>
      ))}
      {/* Solar panel */}
      <mesh position={[0.04, base + poleH * 0.9, 0.06]} rotation={[-0.38, 0, 0.12]}>
        <boxGeometry args={[0.17, 0.007, 0.1]} />
        <meshStandardMaterial color="#1e3a8a" roughness={0.06} metalness={0.82} emissive="#0c1d5e" emissiveIntensity={0.18} />
      </mesh>
      {/* Status orb */}
      <mesh ref={orbRef} position={[0, base + poleH + 0.13, 0]}>
        <sphereGeometry args={[0.062, 10, 8]} />
        <meshStandardMaterial color="#3fb950" emissive="#3fb950" emissiveIntensity={0.55} />
      </mesh>
    </group>
  )
}

// ── Alert beacon ──────────────────────────────────────────────────────────────
function AlertBeacon({ severity, halfH, elevation }) {
  const orbRef   = useRef()
  const lightRef = useRef()
  const color    = severity === 'critical' ? '#f85149' : '#d29922'
  const freq     = severity === 'critical' ? 4.5 : 2.8
  const poleH    = 2.5

  useFrame(({ clock }) => {
    const t = clock.getElapsedTime()
    if (orbRef.current)   orbRef.current.scale.setScalar(0.78 + Math.sin(t * freq) * 0.22)
    if (lightRef.current) lightRef.current.intensity = 0.9 + Math.sin(t * 3.5) * 0.6
  })

  return (
    <group>
      <mesh position={[0, elevation + halfH + poleH / 2, 0]}>
        <cylinderGeometry args={[0.019, 0.024, poleH, 6]} />
        <meshStandardMaterial color="#374151" roughness={0.25} metalness={0.9} />
      </mesh>
      <mesh ref={orbRef} position={[0, elevation + halfH + poleH + 0.22, 0]}>
        <sphereGeometry args={[0.21, 14, 12]} />
        <meshStandardMaterial color={color} emissive={color} emissiveIntensity={3.2} transparent opacity={0.9} />
      </mesh>
      <pointLight ref={lightRef} color={color} intensity={0.9} distance={7}
        position={[0, elevation + halfH + poleH + 0.22, 0]} />
    </group>
  )
}

// ── Main FieldPlot ────────────────────────────────────────────────────────────
export default function FieldPlot({ field, layout, readings, alerts, selected, onSelectField }) {
  const [hovered, setHovered] = useState(false)
  const groupRef   = useRef()
  const topMatRef  = useRef()
  const selRingRef = useRef()
  const liftY      = useRef(0)

  const { x, z, w, d, elevation } = layout
  const halfH = 0.14

  const r = readings ?? {}
  const moisture    = r['soil.moisture']
  const nitrogen    = r['soil.nitrogen']
  const temperature = r['weather.temperature']
  const humidity    = r['weather.humidity']
  const rainfall    = r['weather.rainfall']
  const ph          = r['soil.ph']
  const windSpeed   = r['weather.wind_speed']

  const topColor  = moistureColor(moisture)
  const ringColor = nitrogenColor(nitrogen)
  const critAlert = alerts.find(a => a.severity === 'critical')
  const warnAlert = alerts.find(a => a.severity === 'warning')
  const topAlert  = critAlert ?? warnAlert

  const towerPos = useMemo(() => {
    const mx = w * 0.27, mz = d * 0.27
    return [[-mx, -mz], [mx, -mz], [-mx, mz], [mx, mz]]
  }, [w, d])

  useFrame(() => {
    if (topMatRef.current)
      topMatRef.current.color.lerp(new THREE.Color(topColor), 0.05)
    if (groupRef.current) {
      liftY.current += ((hovered ? 0.22 : 0) - liftY.current) * 0.12
      groupRef.current.position.y = liftY.current
    }
    if (selRingRef.current) selRingRef.current.rotation.z += 0.007
  })

  return (
    <group position={[x, 0, z]}>
      <group ref={groupRef}>

        {/* ── Multi-material slab: earthy brown sides + moisture-coloured top ── */}
        <mesh
          position={[0, elevation + halfH, 0]}
          castShadow receiveShadow
          onPointerOver={e => { e.stopPropagation(); setHovered(true) }}
          onPointerOut={() => setHovered(false)}
          onClick={e => { e.stopPropagation(); onSelectField(field) }}
        >
          <boxGeometry args={[w, halfH * 2, d]} />
          <meshStandardMaterial attach="material-0" color="#5c3a1e" roughness={0.88} />
          <meshStandardMaterial attach="material-1" color="#5c3a1e" roughness={0.88} />
          <meshStandardMaterial attach="material-2" ref={topMatRef} color={topColor} roughness={0.7} metalness={0.04} />
          <meshStandardMaterial attach="material-3" color="#3a2208" roughness={1} />
          <meshStandardMaterial attach="material-4" color="#5c3a1e" roughness={0.88} />
          <meshStandardMaterial attach="material-5" color="#5c3a1e" roughness={0.88} />
        </mesh>

        {/* Elevated terrain plinth (Dry Ridge) */}
        {elevation > 0 && (
          <mesh position={[0, elevation / 2, 0]}>
            <boxGeometry args={[w + 0.1, elevation, d + 0.1]} />
            <meshStandardMaterial color="#2d1e10" roughness={0.95} />
          </mesh>
        )}

        {/* ── Visible crop row furrows on top surface ── */}
        <CropFurrows cropType={field.crop_type} w={w} d={d} elevation={elevation} halfH={halfH} />

        {/* ── Crop geometry ── */}
        <CropInstances cropType={field.crop_type} w={w} d={d} elevation={elevation} halfH={halfH} />

        {/* ── Soil probe stakes ── */}
        <SoilProbes w={w} d={d} elevation={elevation} halfH={halfH} />

        {/* ── IoT sensor towers ── */}
        {towerPos.map(([tx, tz], i) => (
          <SensorTower key={i} tx={tx} tz={tz} elevation={elevation} halfH={halfH} idx={i} />
        ))}

        {/* ── Perimeter fence ── */}
        <FieldFence w={w} d={d} elevation={elevation} />

        {/* Field entrance gate — front centre */}
        <group position={[0, 0, -d / 2]}>
          {/* Left gate post */}
          <mesh position={[-0.35, elevation + 0.32, 0]}>
            <cylinderGeometry args={[0.035, 0.038, 0.64, 5]} />
            <meshLambertMaterial color="#5d4037" />
          </mesh>
          {/* Right gate post */}
          <mesh position={[0.35, elevation + 0.32, 0]}>
            <cylinderGeometry args={[0.035, 0.038, 0.64, 5]} />
            <meshLambertMaterial color="#5d4037" />
          </mesh>
          {/* Horizontal top bar */}
          <mesh position={[0, elevation + 0.56, 0]}>
            <boxGeometry args={[0.72, 0.04, 0.04]} />
            <meshLambertMaterial color="#6d4c2a" />
          </mesh>
          {/* Mid bar */}
          <mesh position={[0, elevation + 0.34, 0]}>
            <boxGeometry args={[0.7, 0.03, 0.03]} />
            <meshLambertMaterial color="#6d4c2a" />
          </mesh>
        </group>

        {/* ── Border trees ── */}
        <BorderTrees w={w} d={d} elevation={elevation} />

        {/* ── Windmill — Sandveld Millet Ridge only (DR) ── */}
        {field.zone_code === 'DR' && (
          <Windmill elevation={elevation} halfH={halfH} windSpeed={windSpeed} w={w} />
        )}

        {/* ── Nitrogen ring ── */}
        {nitrogen != null && (
          <mesh position={[0, elevation + halfH * 2 + 0.03, 0]} rotation={[-Math.PI / 2, 0, 0]}>
            <torusGeometry args={[w * 0.46, 0.032, 8, 52]} />
            <meshStandardMaterial color={ringColor} emissive={ringColor} emissiveIntensity={0.95} />
          </mesh>
        )}

        {/* ── Selection ring — slim, rotating ── */}
        {selected && (
          <mesh ref={selRingRef} position={[0, elevation + halfH * 2 + 0.28, 0]} rotation={[-Math.PI / 2, 0, 0]}>
            <torusGeometry args={[w * 0.47, 0.034, 6, 60]} />
            <meshStandardMaterial color="#58a6ff" emissive="#58a6ff" emissiveIntensity={1.8} transparent opacity={0.88} />
          </mesh>
        )}

        {/* ── Alert beacon ── */}
        {topAlert && <AlertBeacon severity={topAlert.severity} halfH={halfH} elevation={elevation} />}

        {/* ── Rain particles ── */}
        {rainfall != null && rainfall > 1.5 && (
          <RainParticles rainfall={rainfall} width={w} depth={d} x={0} z={0} elevation={elevation + halfH * 2} />
        )}

        {/* ── Always-visible field label ── */}
        <Html position={[0, elevation + halfH * 2 + 3.8, 0]} center distanceFactor={14} style={{ pointerEvents: 'none' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 5, color: '#e6edf3', fontSize: 11,
            fontWeight: 600, whiteSpace: 'nowrap', letterSpacing: '0.25px', textShadow: '0 1px 6px rgba(0,0,0,0.95)' }}>
            <span style={{ width: 7, height: 7, borderRadius: '50%', background: topColor,
              flexShrink: 0, boxShadow: `0 0 5px ${topColor}`, display: 'inline-block' }} />
            {field.name}
            {alerts.length > 0 && (
              <span style={{ color: topAlert?.severity === 'critical' ? '#f85149' : '#d29922', fontWeight: 700 }}>
                ▲{alerts.length}
              </span>
            )}
          </div>
        </Html>

        {/* ── Hover tooltip ── */}
        {hovered && (
          <Html position={[0, elevation + halfH * 2 + 2.0, 0]} center distanceFactor={10} style={{ pointerEvents: 'none' }}>
            <div style={{ background: '#0d1117f2', border: '1px solid #30363d', borderRadius: 8,
              padding: '9px 12px', width: 186, fontSize: 10.5, color: '#e6edf3',
              boxShadow: '0 8px 28px rgba(0,0,0,0.8)', backdropFilter: 'blur(8px)' }}>
              <div style={{ fontWeight: 700, marginBottom: 3, fontSize: 12.5, color: '#58a6ff' }}>{field.name}</div>
              <div style={{ color: '#8b949e', fontSize: 9.5, marginBottom: 7 }}>{field.crop_type} · {field.hectares} ha</div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr auto', gap: '3px 12px' }}>
                {moisture    != null && <GridRow label="Moisture" value={`${moisture}%`}       color={moistureColor(moisture)} />}
                {temperature != null && <GridRow label="Temp"     value={`${temperature}°C`} />}
                {humidity    != null && <GridRow label="Humidity" value={`${humidity}%`} />}
                {nitrogen    != null && <GridRow label="Nitrogen" value={`${nitrogen} mg/kg`}  color={nitrogenColor(nitrogen)} />}
                {ph          != null && <GridRow label="pH"       value={ph} />}
                {rainfall    != null && <GridRow label="Rainfall" value={`${rainfall} mm`}     color="#90caf9" />}
                {windSpeed   != null && <GridRow label="Wind"     value={`${windSpeed} km/h`} />}
              </div>
              {alerts.length > 0 && (
                <div style={{ marginTop: 7, paddingTop: 6, borderTop: '1px solid #21262d' }}>
                  {alerts.slice(0, 3).map(a => (
                    <div key={a.id} style={{ color: a.severity === 'critical' ? '#f85149' : '#d29922', fontSize: 9.5 }}>
                      ⚠ {a.metric}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </Html>
        )}

      </group>
    </group>
  )
}

function GridRow({ label, value, color }) {
  return (
    <>
      <span style={{ color: '#8b949e' }}>{label}</span>
      <span style={{ fontWeight: 600, color: color ?? '#e6edf3', textAlign: 'right' }}>{value}</span>
    </>
  )
}
