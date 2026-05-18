import { memo } from 'react'

interface ChessBoardProps {
  fen: string
  interactive?: boolean
  onSquareClick?: (square: string) => void
}

const PIECE_UNICODE: Record<string, string> = {
  'wK': '♔', 'wQ': '♕', 'wR': '♖', 'wB': '♗', 'wN': '♘', 'wP': '♙',
  'bK': '♚', 'bQ': '♛', 'bR': '♜', 'bB': '♝', 'bN': '♞', 'bP': '♟',
}

function parseFEN(fen: string): string[][] {
  const [position] = fen.split(' ')
  const rows = position.split('/')
  
  return rows.map(row => {
    const squares: string[] = []
    for (const char of row) {
      if (/\d/.test(char)) {
        for (let i = 0; i < parseInt(char); i++) {
          squares.push('')
        }
      } else {
        const color = char === char.toUpperCase() ? 'w' : 'b'
        const piece = char.toUpperCase()
        squares.push(color + piece)
      }
    }
    return squares
  })
}

export const ChessBoard = memo(function ChessBoard({ fen, interactive = false, onSquareClick }: ChessBoardProps) {
  const board = parseFEN(fen)
  const files = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h']
  
  const getSquareName = (row: number, col: number): string => {
    return `${files[col]}${8 - row}`
  }

  return (
    <div className="chess-board">
      {board.map((row, rowIdx) =>
        row.map((piece, colIdx) => {
          const isLight = (rowIdx + colIdx) % 2 === 0
          const squareName = getSquareName(rowIdx, colIdx)
          
          return (
            <div
              key={squareName}
              className={`square ${isLight ? 'light' : 'dark'}`}
              onClick={() => interactive && onSquareClick?.(squareName)}
            >
              {piece && (
                <span className="select-none">
                  {PIECE_UNICODE[piece] || ''}
                </span>
              )}
            </div>
          )
        })
      )}
    </div>
  )
})
