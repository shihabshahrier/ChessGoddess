import { memo, useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'

interface ChessBoardProps {
  fen: string
  interactive?: boolean
  onSquareClick?: (square: string) => void
  lastMove?: { from: string; to: string }
  isBlunder?: boolean
}

const PIECE_UNICODE: Record<string, string> = {
  'wK': '♔', 'wQ': '♕', 'wR': '♖', 'wB': '♗', 'wN': '♘', 'wP': '♙',
  'bK': '♚', 'bQ': '♛', 'bR': '♜', 'bB': '♝', 'bN': '♞', 'bP': '♟',
}

function parseFEN(fen: string): { piece: string; square: string }[][] {
  const [position] = fen.split(' ')
  const rows = position.split('/')
  const files = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h']
  
  return rows.map((row, rowIdx) => {
    const squares: { piece: string; square: string }[] = []
    let colIdx = 0
    
    for (const char of row) {
      if (/\d/.test(char)) {
        for (let i = 0; i < parseInt(char); i++) {
          squares.push({ piece: '', square: `${files[colIdx]}${8 - rowIdx}` })
          colIdx++
        }
      } else {
        const color = char === char.toUpperCase() ? 'w' : 'b'
        const piece = char.toUpperCase()
        squares.push({ 
          piece: color + piece, 
          square: `${files[colIdx]}${8 - rowIdx}` 
        })
        colIdx++
      }
    }
    return squares
  })
}

export const ChessBoard = memo(function ChessBoard({ 
  fen, 
  interactive = false, 
  onSquareClick,
  lastMove,
  isBlunder = false
}: ChessBoardProps) {
  const board = parseFEN(fen)
  const [shake, setShake] = useState(false)

  useEffect(() => {
    if (isBlunder) {
      setShake(true)
      const timer = setTimeout(() => setShake(false), 300)
      return () => clearTimeout(timer)
    }
  }, [isBlunder])

  const isLastMoveSquare = (squareName: string) => {
    if (!lastMove) return false
    return lastMove.from === squareName || lastMove.to === squareName
  }

  return (
    <motion.div 
      className="chess-board"
      animate={shake ? { x: [-2, 2, -2, 2, 0] } : {}}
      transition={{ duration: 0.3 }}
    >
      {board.map((row, rowIdx) =>
        row.map(({ piece, square }, colIdx) => {
          const isLight = (rowIdx + colIdx) % 2 === 0
          const isHighlight = isLastMoveSquare(square)
          
          return (
            <motion.div
              key={square}
              layout
              className={`square ${isLight ? 'light' : 'dark'} ${isHighlight ? 'highlight' : ''}`}
              onClick={() => interactive && onSquareClick?.(square)}
              whileHover={interactive ? { scale: 1.02 } : {}}
              transition={{ type: 'spring', stiffness: 300, damping: 20 }}
            >
              <AnimatePresence mode="wait">
                {piece && (
                  <motion.span
                    key={`${square}-${piece}`}
                    initial={{ scale: 0.8, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    exit={{ scale: 0.8, opacity: 0 }}
                    transition={{ type: 'spring', stiffness: 200, damping: 15 }}
                    className="select-none"
                  >
                    {PIECE_UNICODE[piece]}
                  </motion.span>
                )}
              </AnimatePresence>
            </motion.div>
          )
        })
      )}
    </motion.div>
  )
})
