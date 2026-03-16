import { useEffect, useRef } from 'react'
import { useThree, useFrame } from '@react-three/fiber'
import { Text } from '@react-three/drei'
import * as THREE from 'three'

const FIELD_LABELS = [
  { pos: [-4.5, 0.08, -3.5], text: 'Field 1 · Maize' },
  { pos: [ 3.5, 0.08, -2.5], text: 'Field 2 · Sorghum' },
  { pos: [-3.5, 0.88,  3.0], text: 'Field 3 · Millet' },
  { pos: [ 3.2, 0.08,  3.2], text: 'Field 4 · Groundnuts' },
]

// ── Moon ──────────────────────────────────────────────────────────────────────
function Moon() {
  return (
    <group position={[-24, 15, -30]}>
      <mesh>
        <sphereGeometry args={[2.4, 20, 20]} />
        <meshStandardMaterial color="#d4d8e0" emissive="#c8d0e0" emissiveIntensity={0.55} />
      </mesh>
      <pointLight color="#8ab4d4" intensity={0.75} distance={120} />
    </group>
  )
}

// ── Clouds ────────────────────────────────────────────────────────────────────
const CLOUD_DEFS = [
  { ix: -14, y: 9.5,  z:  -9, speed: 0.50, scale: 1.20 },
  { ix:  -4, y: 11.0, z: -13, speed: 0.35, scale: 0.90 },
  { ix:   8, y: 10.2, z:  -5, speed: 0.45, scale: 1.05 },
  { ix: -20, y:  8.8, z:   5, speed: 0.55, scale: 0.85 },
  { ix:   3, y: 12.0, z:  11, speed: 0.30, scale: 1.10 },
]

function Cloud({ ix, y, z, speed, scale }) {
  const ref = useRef()
  useFrame(({ clock }) => {
    if (ref.current)
      ref.current.position.x = ix + ((clock.getElapsedTime() * speed) % 46) - 2
  })
  return (
    <group ref={ref} position={[ix, y, z]} scale={[scale, scale * 0.52, scale]}>
      <mesh>
        <sphereGeometry args={[0.82, 7, 5]} />
        <meshLambertMaterial color="#18253a" transparent opacity={0.72} />
      </mesh>
      <mesh position={[1.05, -0.12, 0.2]}>
        <sphereGeometry args={[0.66, 7, 5]} />
        <meshLambertMaterial color="#18253a" transparent opacity={0.65} />
      </mesh>
      <mesh position={[-0.9, -0.15, -0.15]}>
        <sphereGeometry args={[0.58, 7, 5]} />
        <meshLambertMaterial color="#1c2840" transparent opacity={0.60} />
      </mesh>
    </group>
  )
}

// ── Birds ─────────────────────────────────────────────────────────────────────
const BIRD_DEFS = Array.from({ length: 12 }, (_, i) => ({
  radius: 8   + (i % 3) * 3.5,
  height: 5   + (i % 4) * 1.6,
  speed:  0.18 + (i % 5) * 0.04,
  phase:  (i / 12) * Math.PI * 2,
}))

function Bird({ radius, height, speed, phase }) {
  const groupRef = useRef()
  const lwRef    = useRef()
  const rwRef    = useRef()

  useFrame(({ clock }) => {
    const t     = clock.getElapsedTime()
    const angle = t * speed + phase
    if (groupRef.current) {
      groupRef.current.position.set(
        Math.cos(angle) * radius,
        height + Math.sin(t * 1.4 + phase) * 0.28,
        Math.sin(angle) * radius,
      )
      groupRef.current.rotation.y = -(angle + Math.PI / 2)
    }
    const flap = Math.sin(t * 3.8 + phase) * 0.38
    if (lwRef.current) lwRef.current.rotation.z =  flap
    if (rwRef.current) rwRef.current.rotation.z = -flap
  })

  return (
    <group ref={groupRef}>
      <mesh>
        <boxGeometry args={[0.07, 0.012, 0.18]} />
        <meshLambertMaterial color="#1e2a3a" />
      </mesh>
      <group ref={lwRef} position={[-0.04, 0, 0]}>
        <mesh position={[-0.12, 0, 0]}>
          <boxGeometry args={[0.22, 0.008, 0.07]} />
          <meshLambertMaterial color="#1e2a3a" />
        </mesh>
      </group>
      <group ref={rwRef} position={[0.04, 0, 0]}>
        <mesh position={[0.12, 0, 0]}>
          <boxGeometry args={[0.22, 0.008, 0.07]} />
          <meshLambertMaterial color="#1e2a3a" />
        </mesh>
      </group>
    </group>
  )
}

// ── Farm infrastructure (near Kavango Maize Belt NB: x=-4.5, z=-3.5) ─────────
function FarmInfrastructure() {
  return (
    <group>
      {/* Water tower */}
      <group position={[-9.4, 0, -5.4]}>
        {[[-0.33,-0.33],[0.33,-0.33],[-0.33,0.33],[0.33,0.33]].map(([lx,lz], i) => (
          <mesh key={i} position={[lx, 0.8, lz]}>
            <cylinderGeometry args={[0.028, 0.032, 1.6, 5]} />
            <meshStandardMaterial color="#546e7a" roughness={0.28} metalness={0.82} />
          </mesh>
        ))}
        <mesh position={[0, 0.55, 0]} rotation={[0, Math.PI / 4, 0]}>
          <boxGeometry args={[0.88, 0.018, 0.018]} />
          <meshStandardMaterial color="#607d8b" roughness={0.35} metalness={0.75} />
        </mesh>
        <mesh position={[0, 0.95, 0]}>
          <boxGeometry args={[0.64, 0.018, 0.64]} />
          <meshStandardMaterial color="#607d8b" roughness={0.35} metalness={0.75} />
        </mesh>
        <mesh position={[0, 1.7, 0]}>
          <cylinderGeometry args={[0.52, 0.52, 0.92, 12]} />
          <meshStandardMaterial color="#78909c" roughness={0.16} metalness={0.72} />
        </mesh>
        <mesh position={[0, 2.2, 0]}>
          <coneGeometry args={[0.58, 0.44, 12]} />
          <meshStandardMaterial color="#546e7a" roughness={0.24} metalness={0.62} />
        </mesh>
        <mesh position={[0.48, 1.2, 0]}>
          <cylinderGeometry args={[0.03, 0.03, 0.75, 5]} />
          <meshStandardMaterial color="#90a4ae" roughness={0.28} metalness={0.8} />
        </mesh>
      </group>

      {/* Storage shed */}
      <group position={[-8.8, 0, -2.0]}>
        <mesh position={[0, 0.54, 0]}>
          <boxGeometry args={[1.7, 1.08, 1.05]} />
          <meshLambertMaterial color="#8d6e63" />
        </mesh>
        <mesh position={[0, 1.16,  0.24]} rotation={[-0.44, 0, 0]}>
          <boxGeometry args={[1.82, 0.065, 0.74]} />
          <meshLambertMaterial color="#5d4037" />
        </mesh>
        <mesh position={[0, 1.16, -0.24]} rotation={[0.44, 0, 0]}>
          <boxGeometry args={[1.82, 0.065, 0.74]} />
          <meshLambertMaterial color="#5d4037" />
        </mesh>
        <mesh position={[0, 0.3, 0.535]}>
          <boxGeometry args={[0.34, 0.6, 0.03]} />
          <meshLambertMaterial color="#4e342e" />
        </mesh>
        <mesh position={[0.58, 0.68, 0.535]}>
          <boxGeometry args={[0.22, 0.22, 0.03]} />
          <meshLambertMaterial color="#1565c0" />
        </mesh>
      </group>
    </group>
  )
}

export default function SceneExtras() {
  const { scene } = useThree()

  useEffect(() => {
    scene.background = new THREE.Color('#080d12')
    scene.fog = new THREE.FogExp2('#080d12', 0.011)
  }, [scene])

  return (
    <>
      <hemisphereLight skyColor="#2a4a6e" groundColor="#1a2e1a" intensity={0.55} />
      <directionalLight
        position={[12, 22, 8]} intensity={1.65} color="#fff8e1" castShadow
        shadow-mapSize-width={2048} shadow-mapSize-height={2048}
        shadow-camera-far={60} shadow-camera-left={-22} shadow-camera-right={22}
        shadow-camera-top={22} shadow-camera-bottom={-22}
      />
      <directionalLight position={[-8, 10, -6]} intensity={0.38} color="#7eb8d4" />
      <pointLight position={[0, 5, -12]} intensity={0.45} color="#ff8a50" distance={25} />

      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.02, 0]} receiveShadow>
        <planeGeometry args={[80, 80, 1, 1]} />
        <meshLambertMaterial color="#7a6535" />
      </mesh>
      <gridHelper args={[80, 80, '#6b5828', '#5a4a20']} position={[0, -0.01, 0]} />

      <Moon />
      {CLOUD_DEFS.map((c, i) => <Cloud key={i} {...c} />)}
      {BIRD_DEFS.map((b, i) => <Bird key={i} {...b} />)}
      <FarmInfrastructure />

      {/* Trees scattered around and between the fields */}
      {[
        // North edge / corners
        [-9.5,-8.5],[-8.2,-8.0],[-6.8,-7.6],[-5.1,-8.2],[-3.4,-7.8],[-1.8,-8.4],
        [ 0.2,-8.1],[ 1.9,-7.5],[ 3.6,-8.3],[ 5.2,-7.9],[ 6.8,-8.5],[ 8.4,-7.7],
        // South edge
        [-8.8, 9.2],[-7.1, 8.6],[-5.4, 9.5],[-3.6, 8.9],[-1.8, 9.3],[0.4, 8.7],
        [ 2.2, 9.4],[ 4.1, 8.8],[ 5.9, 9.1],[ 7.6, 8.5],[ 9.0, 9.3],
        // West edge
        [-10.8,-5.5],[-11.2,-3.2],[-10.5,-0.8],[-11.0, 1.8],[-10.6, 4.5],[-11.1, 6.8],
        // East edge
        [ 10.2,-6.0],[ 11.0,-3.5],[ 10.5,-1.0],[ 11.2, 1.5],[ 10.7, 4.2],[ 11.1, 6.9],
        // Gaps between fields (natural scrub clusters)
        [-1.6,-1.8],[-0.8,-3.0],[-0.3,-0.5],[-1.2, 0.8],[-0.5, 2.2],
        [ 1.2,-1.4],[ 0.6, 0.2],[ 1.5, 1.6],[ 0.4,-4.2],[ 1.8,-3.5],
        // Loose scatter far corners
        [-13.0, 0.5],[-12.5, 3.8],[-13.2,-3.0],[ 12.8, 0.0],[ 12.4,-3.8],[ 12.9, 3.5],
        [-9.0, 7.5],[ 9.5, 7.0],[-7.5,-9.8],[ 7.8,-9.5],
      ].map(([tx, tz], i) => {
        const h = 0.5 + Math.sin(i * 1.7 + 0.4) * 0.28
        const w = 0.24 + Math.sin(i * 3.1 + 1.1) * 0.07
        return (
          <group key={i} position={[tx, 0, tz]}>
            <mesh position={[0, h * 0.3, 0]} castShadow>
              <cylinderGeometry args={[0.035, 0.055, h * 0.6, 5]} />
              <meshLambertMaterial color="#3e2a1a" />
            </mesh>
            <mesh position={[0, h * 1.1, 0]} castShadow>
              <coneGeometry args={[w + 0.04, h * 1.1, 5]} />
              <meshLambertMaterial color="#0d1f0d" />
            </mesh>
            <mesh position={[0, h * 1.65, 0]} castShadow>
              <coneGeometry args={[w * 0.65, h * 0.7, 5]} />
              <meshLambertMaterial color="#112811" />
            </mesh>
          </group>
        )
      })}


{FIELD_LABELS.map((l, i) => (
        <Text key={i} position={[l.pos[0], l.pos[1], l.pos[2]]} rotation={[-Math.PI / 2, 0, 0]}
          fontSize={0.22} color="#c9d1d9" anchorX="center" anchorY="middle"
          letterSpacing={0.02} outlineWidth={0.015} outlineColor="#000000">
          {l.text}
        </Text>
      ))}

      <Text position={[0, 0.05, -15]} rotation={[-Math.PI / 2, 0, 0]}
        fontSize={0.5} color="#3fb950" anchorX="center" outlineWidth={0.02} outlineColor="#000">
        ↑ N
      </Text>
    </>
  )
}
