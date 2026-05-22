import { memo } from 'react'
import { motion } from 'framer-motion'

interface EvalBarProps {
  evaluation: number // pawns, White POV
  mate?: number // White POV moves-to-mate; 0/undefined = none
  height?: number // fixed px height; omit to stretch to the parent
}

export const EvalBar = memo(function EvalBar({
  evaluation,
  mate = 0,
  height,
}: EvalBarProps) {
  const whitePercent =
    mate !== 0
      ? mate > 0
        ? 100
        : 0
      : ((Math.max(-10, Math.min(10, evaluation)) + 10) / 20) * 100

  const label =
    mate !== 0
      ? `M${Math.abs(mate)}`
      : Math.abs(evaluation) >= 50 // mate sentinel stored as a large pawn value
        ? evaluation > 0
          ? '+M'
          : '−M'
        : `${evaluation > 0 ? '+' : ''}${evaluation.toFixed(1)}`

  return (
    <div
      className={`relative w-8 rounded-lg overflow-hidden border border-chess-border bg-gray-800 ${
        height ? '' : 'self-stretch'
      }`}
      style={height ? { height } : undefined}
    >
      <motion.div
        className="absolute bottom-0 left-0 right-0 bg-white"
        animate={{ height: `${whitePercent}%` }}
        transition={{ type: 'spring', stiffness: 120, damping: 16, mass: 0.8 }}
      />
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <span className="text-[10px] font-mono text-white mix-blend-difference font-semibold">
          {label}
        </span>
      </div>
    </div>
  )
})
