import { useRef, useCallback, useEffect, useState } from 'react'
import { motion } from 'framer-motion'

interface Move {
  id: string
  moveNumber: number
  san: string
  classification?: 'blunder' | 'mistake' | 'inaccuracy' | 'good' | 'excellent' | 'best'
  evaluation?: number
}

interface TimelineProps {
  moves: Move[]
  currentMoveIndex: number
  onMoveSelect: (index: number) => void
}

const classificationColors: Record<string, string> = {
  blunder: 'bg-red-500',
  mistake: 'bg-orange-500',
  inaccuracy: 'bg-yellow-500',
  good: 'bg-green-500',
  excellent: 'bg-chess-gold',
  best: 'bg-chess-gold',
}

export function Timeline({ moves, currentMoveIndex, onMoveSelect }: TimelineProps) {
  const timelineRef = useRef<HTMLDivElement>(null)
  const [isDragging, setIsDragging] = useState(false)

  const handleScroll = useCallback(() => {
    if (!timelineRef.current || isDragging) return
    
    const scrollLeft = timelineRef.current.scrollLeft
    const scrollWidth = timelineRef.current.scrollWidth - timelineRef.current.clientWidth
    const scrollPercent = scrollLeft / scrollWidth
    
    const moveIndex = Math.round(scrollPercent * (moves.length - 1))
    onMoveSelect(Math.max(0, Math.min(moves.length - 1, moveIndex)))
  }, [isDragging, moves.length, onMoveSelect])

  const handleWheel = useCallback((e: React.WheelEvent) => {
    if (!timelineRef.current) return
    e.preventDefault()
    timelineRef.current.scrollLeft += e.deltaY
  }, [])

  useEffect(() => {
    if (!timelineRef.current || isDragging) return
    
    const scrollWidth = timelineRef.current.scrollWidth - timelineRef.current.clientWidth
    const targetScroll = (currentMoveIndex / (moves.length - 1)) * scrollWidth
    timelineRef.current.scrollLeft = targetScroll
  }, [currentMoveIndex, moves.length, isDragging])

  return (
    <div className="w-full">
      <div 
        ref={timelineRef}
        onScroll={handleScroll}
        onWheel={handleWheel}
        className="flex gap-1 overflow-x-auto pb-2 scrollbar-thin scrollbar-thumb-chess-border scrollbar-track-chess-surface"
        style={{ scrollbarWidth: 'thin' }}
      >
        {moves.map((move, idx) => (
          <motion.button
            key={move.id}
            onClick={() => onMoveSelect(idx)}
            onMouseDown={() => setIsDragging(true)}
            onMouseUp={() => setIsDragging(false)}
            className={`flex-shrink-0 w-12 h-16 rounded-lg border transition-all ${
              idx === currentMoveIndex 
                ? 'border-chess-gold bg-chess-gold/10' 
                : 'border-chess-border bg-chess-surface hover:border-chess-gold/50'
            }`}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            <div className="flex flex-col items-center justify-center h-full">
              <span className="text-xs text-chess-text-dim">{move.moveNumber}</span>
              <span className="text-sm font-mono text-chess-text">{move.san}</span>
              {move.classification && (
                <div className={`w-2 h-2 rounded-full mt-1 ${classificationColors[move.classification]}`} />
              )}
            </div>
          </motion.button>
        ))}
      </div>
      
      <div className="mt-2 text-center text-xs text-chess-text-dim">
        Scroll to scrub through moves
      </div>
    </div>
  )
}
