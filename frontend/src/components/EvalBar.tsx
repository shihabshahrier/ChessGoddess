import { memo } from 'react'

interface EvalBarProps {
  evaluation: number
  height?: number
}

export const EvalBar = memo(function EvalBar({ evaluation, height = 400 }: EvalBarProps) {
  // Convert evaluation to percentage (cap at ±10)
  const cappedEval = Math.max(-10, Math.min(10, evaluation))
  const whitePercent = ((cappedEval + 10) / 20) * 100

  return (
    <div 
      className="eval-bar rounded-lg overflow-hidden border border-chess-border"
      style={{ height: `${height}px` }}
    >
      <div 
        className="eval-bar-white"
        style={{ 
          height: `${whitePercent}%`,
          transition: 'height 0.6s cubic-bezier(0.34, 1.56, 0.64, 1)'
        }}
      />
      <div className="absolute bottom-2 left-0 right-0 text-center">
        <span className="text-xs font-mono text-chess-text bg-chess-bg/80 px-1 rounded">
          {evaluation > 0 ? '+' : ''}{evaluation.toFixed(2)}
        </span>
      </div>
    </div>
  )
})
