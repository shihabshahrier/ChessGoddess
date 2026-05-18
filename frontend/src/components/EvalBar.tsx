import { memo } from 'react'
import { motion } from 'framer-motion'

interface EvalBarProps {
  evaluation: number
  height?: number
}

export const EvalBar = memo(function EvalBar({ evaluation, height = 400 }: EvalBarProps) {
  const cappedEval = Math.max(-10, Math.min(10, evaluation))
  const whitePercent = ((cappedEval + 10) / 20) * 100

  return (
    <div 
      className="relative rounded-lg overflow-hidden border border-chess-border bg-gray-800"
      style={{ height: `${height}px`, width: '32px' }}
    >
      <motion.div 
        className="absolute bottom-0 left-0 right-0 bg-white"
        style={{ height: `${whitePercent}%` }}
        animate={{ height: `${whitePercent}%` }}
        transition={{ 
          type: 'spring', 
          stiffness: 120, 
          damping: 14,
          mass: 0.8 
        }}
      />
      
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <span className="text-xs font-mono text-white mix-blend-difference font-medium">
          {evaluation > 0 ? '+' : ''}{evaluation.toFixed(1)}
        </span>
      </div>
    </div>
  )
})
