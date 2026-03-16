import { useRef, useMemo } from 'react'
import { useFrame } from '@react-three/fiber'
import * as THREE from 'three'

export default function RainParticles({ rainfall, width, depth, x, z, elevation }) {
  const count = Math.min(Math.floor(rainfall * 8), 60)
  const meshRef = useRef()

  // Stable random initial positions
  const positions = useMemo(() => {
    const arr = new Float32Array(count * 3)
    for (let i = 0; i < count; i++) {
      arr[i * 3]     = x + (Math.random() - 0.5) * width
      arr[i * 3 + 1] = elevation + Math.random() * 3
      arr[i * 3 + 2] = z + (Math.random() - 0.5) * depth
    }
    return arr
  }, [count, x, z, width, depth, elevation])

  useFrame(() => {
    if (!meshRef.current) return
    const pos = meshRef.current.geometry.attributes.position
    for (let i = 0; i < count; i++) {
      pos.array[i * 3 + 1] -= 0.07
      // Reset to top when below ground
      if (pos.array[i * 3 + 1] < elevation - 0.1) {
        pos.array[i * 3]     = x + (Math.random() - 0.5) * width
        pos.array[i * 3 + 1] = elevation + 3
        pos.array[i * 3 + 2] = z + (Math.random() - 0.5) * depth
      }
    }
    pos.needsUpdate = true
  })

  if (count === 0) return null

  return (
    <points ref={meshRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          args={[positions, 3]}
        />
      </bufferGeometry>
      <pointsMaterial
        color="#90caf9"
        size={0.07}
        transparent
        opacity={0.8}
        sizeAttenuation
      />
    </points>
  )
}
