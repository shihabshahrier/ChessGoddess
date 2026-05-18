import { memo } from 'react'

interface Move {
  id: string
  moveNumber: number
  san: string
  classification?: 'blunder' | 'mistake' | 'inaccuracy' | 'good' | 'excellent' | 'best'
  evaluation?: number
}

interface MoveListProps {
  moves: Move[]
  interactive?: boolean
  onMoveClick?: (moveId: string) => void
}

const classificationColors: Record<string, string> = {
  blunder: 'text-red-500',
  mistake: 'text-orange-500',
  inaccuracy: 'text-yellow-500',
  good: 'text-green-500',
  excellent: 'text-chess-gold',
  best: 'text-chess-gold font-bold',
}

const classificationIcons: Record<string, string> = {
  blunder: '??',
  mistake: '?',
  inaccuracy: '?!',
  good: '!',
  excellent: '!!',
  best: '★',
}

export const MoveList = memo(function MoveList({ moves, interactive = false, onMoveClick }: MoveListProps) {
  const movePairs: { white?: Move; black?: Move }[] = []
  
  for (let i = 0; i < moves.length; i += 2) {
    movePairs.push({
      white: moves[i],
      black: moves[i + 1],
    })
  }

  if (moves.length === 0) {
    return (
      <div className="text-chess-text-muted text-center py-8">
        No moves to display
      </div>
    )
  }

  return (
    <div className="space-y-1 font-mono text-sm">
      {movePairs.map((pair, idx) => (
        <div key={idx} className="flex items-center gap-2">
          <span className="text-chess-text-dim w-8">{idx + 1}.</span>
          
          {pair.white && (
            <button
              className={`flex items-center gap-1 px-2 py-1 rounded hover:bg-chess-elevated transition-colors flex-1 justify-start ${
                interactive ? 'cursor-pointer' : 'cursor-default'
              }`}
              onClick={() => interactive && onMoveClick?.(pair.white!.id)}
            >
              <span className="text-chess-text">{pair.white.san}</span>
              {pair.white.classification && (
                <span className={`text-xs ${classificationColors[pair.white.classification]}`}>
                  {classificationIcons[pair.white.classification]}
                </span>
              )}
            </button>
          )}
          
          {pair.black && (
            <button
              className={`flex items-center gap-1 px-2 py-1 rounded hover:bg-chess-elevated transition-colors flex-1 justify-start ${
                interactive ? 'cursor-pointer' : 'cursor-default'
              }`}
              onClick={() => interactive && onMoveClick?.(pair.black!.id)}
            >
              <span className="text-chess-text">{pair.black.san}</span>
              {pair.black.classification && (
                <span className={`text-xs ${classificationColors[pair.black.classification]}`}>
                  {classificationIcons[pair.black.classification]}
                </span>
              )}
            </button>
          )}
        </div>
      ))}
    </div>
  )
})
