import { memo, useRef } from 'react'

interface EvalGraphProps {
  evaluations: number[] // White POV pawns, one per move
  currentIndex: number // 0 = start, k = after move k
  onSelect: (index: number) => void
}

const H = 48
const CAP = 6 // clamp eval display to ±6 pawns

export const EvalGraph = memo(function EvalGraph({
  evaluations,
  currentIndex,
  onSelect,
}: EvalGraphProps) {
  const ref = useRef<SVGSVGElement>(null)
  const n = evaluations.length
  if (n === 0) return null

  const points = n + 1 // includes the start position
  const W = points - 1

  const yOf = (e: number) => {
    const c = Math.max(-CAP, Math.min(CAP, e))
    return H / 2 - (c / CAP) * (H / 2)
  }
  const evalAt = (i: number) => (i === 0 ? 0 : evaluations[i - 1])

  const linePts = Array.from(
    { length: points },
    (_, i) => `${i},${yOf(evalAt(i))}`,
  ).join(' ')
  const areaPts = `0,${H / 2} ${linePts} ${W},${H / 2}`

  const handleClick = (e: React.MouseEvent) => {
    const rect = ref.current?.getBoundingClientRect()
    if (!rect) return
    const ratio = (e.clientX - rect.left) / rect.width
    onSelect(Math.max(0, Math.min(points - 1, Math.round(ratio * W))))
  }

  return (
    <svg
      ref={ref}
      viewBox={`0 0 ${W} ${H}`}
      preserveAspectRatio="none"
      className="w-full h-12 cursor-pointer rounded-md bg-gray-900/40"
      onClick={handleClick}
    >
      <rect x="0" y="0" width={W} height={H / 2} fill="rgba(255,255,255,0.05)" />
      <polyline points={areaPts} fill="rgba(212,175,55,0.18)" stroke="none" />
      <line
        x1="0"
        y1={H / 2}
        x2={W}
        y2={H / 2}
        stroke="rgba(255,255,255,0.18)"
        strokeWidth="1"
        vectorEffect="non-scaling-stroke"
      />
      <polyline
        points={linePts}
        fill="none"
        stroke="var(--chess-gold)"
        strokeWidth="1.5"
        vectorEffect="non-scaling-stroke"
      />
      <line
        x1={currentIndex}
        y1="0"
        x2={currentIndex}
        y2={H}
        stroke="rgba(255,255,255,0.7)"
        strokeWidth="1.5"
        vectorEffect="non-scaling-stroke"
      />
    </svg>
  )
})
