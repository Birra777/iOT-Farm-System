import { useMemo, useRef } from 'react'
import * as THREE from 'three'

function gridPositions(count, w, d, margin = 0.45) {
  const cols = Math.ceil(Math.sqrt(count * (w / d)))
  const rows = Math.ceil(count / cols)
  const out = []
  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      if (out.length >= count) break
      out.push([
        -w / 2 + margin + (c / Math.max(cols - 1, 1)) * (w - margin * 2),
        0,
        -d / 2 + margin + (r / Math.max(rows - 1, 1)) * (d - margin * 2),
      ])
    }
  }
  return out
}

function seededPositions(count, w, d, margin = 0.35, seed = 42) {
  let s = seed
  const rand = () => { s = (s * 1664525 + 1013904223) & 0xffffffff; return (s >>> 0) / 0xffffffff }
  return Array.from({ length: count }, () => [
    -w / 2 + margin + rand() * (w - margin * 2),
    0,
    -d / 2 + margin + rand() * (d - margin * 2),
  ])
}

// ── Maize ── stalk + wide leaf-bundle cone + narrow golden tassel
function MaizeField({ w, d, elevation, halfH }) {
  const count = 36
  const positions = useMemo(() => gridPositions(count, w, d), [w, d])
  const stalkRef  = useRef()
  const leafRef   = useRef()
  const tasselRef = useRef()
  const dummy = useMemo(() => new THREE.Object3D(), [])

  useMemo(() => {
    const base = elevation + halfH * 2
    const layers = [
      {
        ref: stalkRef,
        fn: (x, z, i) => {
          dummy.position.set(x, base + 0.27 + Math.sin(i * 37.1) * 0.05, z)
          dummy.scale.set(1, 0.82 + Math.sin(i * 37.1) * 0.22, 1)
          dummy.rotation.y = i * 1.37
        },
      },
      {
        ref: leafRef,
        fn: (x, z, i) => {
          dummy.position.set(x, base + 0.14, z)
          dummy.scale.set(
            1.0 + Math.sin(i * 13.7) * 0.22,
            0.75 + Math.cos(i * 7.3) * 0.12,
            1.0 + Math.sin(i * 19.1) * 0.22,
          )
          dummy.rotation.y = i * 1.05
        },
      },
      {
        ref: tasselRef,
        fn: (x, z, i) => {
          dummy.position.set(x, base + 0.52 + Math.sin(i * 37.1) * 0.04, z)
          dummy.scale.set(
            0.75 + Math.sin(i * 41.3) * 0.18,
            1.0 + Math.sin(i * 29.1) * 0.28,
            0.75 + Math.sin(i * 41.3) * 0.18,
          )
          dummy.rotation.y = i * 0.9
        },
      },
    ]
    layers.forEach(({ ref, fn }) => {
      if (!ref.current) return
      positions.forEach(([x, , z], i) => {
        fn(x, z, i)
        dummy.updateMatrix()
        ref.current.setMatrixAt(i, dummy.matrix)
      })
      ref.current.instanceMatrix.needsUpdate = true
    })
  })

  return (
    <>
      <instancedMesh ref={stalkRef} args={[null, null, count]}>
        <cylinderGeometry args={[0.026, 0.034, 0.52, 5]} />
        <meshLambertMaterial color="#2d6a1a" />
      </instancedMesh>
      <instancedMesh ref={leafRef} args={[null, null, count]}>
        <coneGeometry args={[0.15, 0.11, 6]} />
        <meshLambertMaterial color="#3d8020" />
      </instancedMesh>
      <instancedMesh ref={tasselRef} args={[null, null, count]}>
        <coneGeometry args={[0.038, 0.2, 5]} />
        <meshLambertMaterial color="#c8b400" />
      </instancedMesh>
    </>
  )
}

// ── Sorghum ── stalk + leaf-bundle + drooping dark-red grain head
function SorghumField({ w, d, elevation, halfH }) {
  const count = 30
  const positions = useMemo(() => gridPositions(count, w, d, 0.48), [w, d])
  const stalkRef = useRef()
  const leafRef  = useRef()
  const headRef  = useRef()
  const dummy = useMemo(() => new THREE.Object3D(), [])

  useMemo(() => {
    if (!stalkRef.current || !leafRef.current || !headRef.current) return
    const base = elevation + halfH * 2
    positions.forEach(([x, , z], i) => {
      const sc = 0.78 + Math.sin(i * 53.1) * 0.17
      dummy.position.set(x, base + 0.20, z)
      dummy.scale.set(sc, sc, sc)
      dummy.rotation.y = i * 2.1
      dummy.updateMatrix()
      stalkRef.current.setMatrixAt(i, dummy.matrix)

      dummy.position.set(x, base + 0.12, z)
      dummy.scale.set(sc * 1.35, sc * 0.7, sc * 1.35)
      dummy.updateMatrix()
      leafRef.current.setMatrixAt(i, dummy.matrix)

      dummy.position.set(
        x + Math.sin(i * 2.3) * 0.07,
        base + 0.44,
        z + Math.cos(i * 1.7) * 0.07,
      )
      dummy.scale.set(sc * 0.88, sc * 1.35, sc * 0.88)
      dummy.rotation.y = i * 1.3
      dummy.updateMatrix()
      headRef.current.setMatrixAt(i, dummy.matrix)
    })
    stalkRef.current.instanceMatrix.needsUpdate = true
    leafRef.current.instanceMatrix.needsUpdate = true
    headRef.current.instanceMatrix.needsUpdate = true
  })

  return (
    <>
      <instancedMesh ref={stalkRef} args={[null, null, count]}>
        <cylinderGeometry args={[0.028, 0.036, 0.38, 5]} />
        <meshLambertMaterial color="#4a7c3f" />
      </instancedMesh>
      <instancedMesh ref={leafRef} args={[null, null, count]}>
        <coneGeometry args={[0.13, 0.1, 6]} />
        <meshLambertMaterial color="#3a6a30" />
      </instancedMesh>
      <instancedMesh ref={headRef} args={[null, null, count]}>
        <sphereGeometry args={[0.08, 7, 6]} />
        <meshLambertMaterial color="#7a2e0a" />
      </instancedMesh>
    </>
  )
}

// ── Millet ── drought-stressed: sparse, yellowed, arching seed plumes
function MilletField({ w, d, elevation, halfH }) {
  const count = 22
  const positions = useMemo(() => seededPositions(count, w, d, 0.3, 7), [w, d])
  const stemRef  = useRef()
  const plumeRef = useRef()
  const dummy = useMemo(() => new THREE.Object3D(), [])

  useMemo(() => {
    if (!stemRef.current || !plumeRef.current) return
    const base = elevation + halfH * 2
    positions.forEach(([x, , z], i) => {
      const sc = 0.55 + Math.sin(i * 17.9) * 0.28
      dummy.position.set(x, base + 0.12, z)
      dummy.scale.set(sc, sc, sc)
      dummy.rotation.y = i * 2.1
      dummy.updateMatrix()
      stemRef.current.setMatrixAt(i, dummy.matrix)

      const arch = Math.sin(i * 3.7) * 0.09
      dummy.position.set(x + arch, base + 0.28, z + arch * 0.6)
      dummy.scale.set(sc * 0.72, sc * 1.5, sc * 0.72)
      dummy.rotation.z = arch * 2.5
      dummy.rotation.y = i * 1.8
      dummy.updateMatrix()
      plumeRef.current.setMatrixAt(i, dummy.matrix)
    })
    stemRef.current.instanceMatrix.needsUpdate = true
    plumeRef.current.instanceMatrix.needsUpdate = true
  })

  return (
    <>
      <instancedMesh ref={stemRef} args={[null, null, count]}>
        <cylinderGeometry args={[0.023, 0.03, 0.24, 5]} />
        <meshLambertMaterial color="#8c9a2a" />
      </instancedMesh>
      <instancedMesh ref={plumeRef} args={[null, null, count]}>
        <sphereGeometry args={[0.075, 7, 5]} />
        <meshLambertMaterial color="#c8a830" />
      </instancedMesh>
    </>
  )
}

// ── Groundnuts ── low spreading bushes, two oblate-sphere layers per plant
function GroundnutField({ w, d, elevation, halfH }) {
  const count = 40
  const positions = useMemo(() => seededPositions(count, w, d, 0.42, 13), [w, d])
  const bodyRef = useRef()
  const lobeRef = useRef()
  const dummy = useMemo(() => new THREE.Object3D(), [])

  useMemo(() => {
    if (!bodyRef.current || !lobeRef.current) return
    const base = elevation + halfH * 2
    positions.forEach(([x, , z], i) => {
      dummy.position.set(x, base + 0.042, z)
      dummy.scale.set(
        0.85 + Math.sin(i * 29.1) * 0.18,
        0.32 + Math.sin(i * 13.7) * 0.08,
        0.85 + Math.sin(i * 41.3) * 0.18,
      )
      dummy.rotation.y = i * 0.85
      dummy.updateMatrix()
      bodyRef.current.setMatrixAt(i, dummy.matrix)

      dummy.position.set(
        x + Math.sin(i * 1.3) * 0.08,
        base + 0.065,
        z + Math.cos(i * 2.1) * 0.08,
      )
      dummy.scale.set(
        0.48 + Math.sin(i * 7.3) * 0.12,
        0.25 + Math.sin(i * 11.1) * 0.06,
        0.48 + Math.sin(i * 7.3) * 0.12,
      )
      dummy.rotation.y = i * 1.4
      dummy.updateMatrix()
      lobeRef.current.setMatrixAt(i, dummy.matrix)
    })
    bodyRef.current.instanceMatrix.needsUpdate = true
    lobeRef.current.instanceMatrix.needsUpdate = true
  })

  return (
    <>
      <instancedMesh ref={bodyRef} args={[null, null, count]}>
        <sphereGeometry args={[0.13, 7, 5]} />
        <meshLambertMaterial color="#6b8e23" />
      </instancedMesh>
      <instancedMesh ref={lobeRef} args={[null, null, count]}>
        <sphereGeometry args={[0.1, 6, 5]} />
        <meshLambertMaterial color="#87a92e" />
      </instancedMesh>
    </>
  )
}

export default function CropInstances({ cropType, w, d, elevation, halfH }) {
  const props = { w, d, elevation, halfH }
  switch (cropType) {
    case 'Maize':      return <MaizeField {...props} />
    case 'Sorghum':    return <SorghumField {...props} />
    case 'Millet':     return <MilletField {...props} />
    case 'Groundnuts': return <GroundnutField {...props} />
    default:           return null
  }
}
