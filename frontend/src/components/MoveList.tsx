import { memo } from 'react'
import type { MoveClassification } from '../types'
import { CLASSIFICATION_META } from '../utils/classification'

export interface MoveListItem {
  id: string
  san: string
  classification?: MoveClassification
}

interface MoveListProps {
  moves: MoveListItem[]
  onMoveClick?: (moveId: string) => void
  activeMoveId?: string
}

export const MoveList = memo(function MoveList({
  moves,
  onMoveClick,
  activeMoveId,
}: MoveListProps) {
  if (moves.length === 0) {
    return (
      <div className="text-chess-text-muted text-center py-8 text-sm">
        No moves to display
      </div>
    )
  }

  const pairs: { white?: MoveListItem; black?: MoveListItem }[] = []
  for (let i = 0; i < moves.length; i += 2) {
    pairs.push({ white: moves[i], black: moves[i + 1] })
  }

  const cell = (move?: MoveListItem) => {
    if (!move) return <span className="flex-1" />
    const meta = move.classification
      ? CLASSIFICATION_META[move.classification]
      : null
    const active = move.id === activeMoveId
    return (
      <button
        onClick={() => onMoveClick?.(move.id)}
        className={`flex flex-1 items-center gap-1 px-2 py-1 rounded transition-colors justify-start ${
          active
            ? 'bg-chess-gold/20 text-chess-gold'
            : 'hover:bg-chess-elevated text-chess-text'
        }`}
      >
        <span>{move.san}</span>
        {meta && <span className={`text-xs ${meta.color}`}>{meta.icon}</span>}
      </button>
    )
  }

  return (
    <div className="space-y-0.5 font-mono text-sm max-h-96 overflow-y-auto">
      {pairs.map((pair, idx) => (
        <div key={idx} className="flex items-center gap-2">
          <span className="text-chess-text-dim w-7 shrink-0">{idx + 1}.</span>
          {cell(pair.white)}
          {cell(pair.black)}
        </div>
      ))}
    </div>
  )
})
