import { memo, useState, useMemo, useCallback } from 'react'
import { Chess } from 'chess.js'
import { motion, AnimatePresence } from 'framer-motion'
import type { LegalMove } from '../hooks/useChessGame'

/** Resolve a piece key ("wk", "bq", …) to its bundled cburnett SVG. */
function pieceSrc(key: string): string {
  return `/pieces/${key[0]}${key[1].toUpperCase()}.svg`
}

const FILES = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h']
const ORDER = [0, 1, 2, 3, 4, 5, 6, 7]
const PROMO_PIECES = ['q', 'r', 'b', 'n'] as const

/** An arrow drawn over the board, e.g. an engine suggestion. */
export interface BoardArrow {
  from: string
  to: string
  kind?: 'primary' | 'secondary'
}

interface InteractiveBoardProps {
  fen: string
  onMove: (from: string, to: string, promotion?: string) => void
  getLegalMoves: (square: string) => LegalMove[]
  orientation?: 'white' | 'black'
  lastMove?: { from: string; to: string } | null
  arrows?: BoardArrow[]
  disabled?: boolean
}

/** Map a square to its centre in board grid units (0-8), respecting orientation. */
function squareToXY(square: string, orientation: 'white' | 'black') {
  const file = square.charCodeAt(0) - 97
  const rank = parseInt(square.slice(1), 10)
  let col = file
  let row = 8 - rank
  if (orientation === 'black') {
    col = 7 - col
    row = 7 - row
  }
  return { x: col + 0.5, y: row + 0.5 }
}

export const InteractiveBoard = memo(function InteractiveBoard({
  fen,
  onMove,
  getLegalMoves,
  orientation = 'white',
  lastMove = null,
  arrows = [],
  disabled = false,
}: InteractiveBoardProps) {
  const [selected, setSelected] = useState<string | null>(null)
  const [pendingPromotion, setPendingPromotion] =
    useState<{ from: string; to: string } | null>(null)

  // Derive board grid, side to move, and king-in-check square from the FEN.
  const { board, turn, checkSquare } = useMemo(() => {
    const g = new Chess(fen)
    const b = g.board()
    let ck: string | null = null
    if (g.inCheck()) {
      for (const row of b)
        for (const cell of row)
          if (cell && cell.type === 'k' && cell.color === g.turn())
            ck = cell.square
    }
    return { board: b, turn: g.turn(), checkSquare: ck }
  }, [fen])

  // Legal destinations for the currently selected piece: square -> flags.
  const legalTargets = useMemo(() => {
    const m = new Map<string, string>()
    if (selected)
      for (const mv of getLegalMoves(selected)) m.set(mv.to, mv.flags)
    return m
  }, [selected, getLegalMoves])

  const tryMove = useCallback(
    (from: string, to: string): boolean => {
      const flags = legalTargets.get(to)
      if (flags === undefined) return false
      if (flags.includes('p')) {
        setPendingPromotion({ from, to }) // pawn reaching last rank
        return true
      }
      onMove(from, to)
      return true
    },
    [legalTargets, onMove],
  )

  const handleSquareClick = useCallback(
    (square: string, hasOwnPiece: boolean) => {
      if (disabled || pendingPromotion) return
      if (selected) {
        if (square === selected) return setSelected(null)
        if (tryMove(selected, square)) return setSelected(null)
        return setSelected(hasOwnPiece ? square : null)
      }
      if (hasOwnPiece) setSelected(square)
    },
    [disabled, pendingPromotion, selected, tryMove],
  )

  const handleDrop = useCallback(
    (square: string) => {
      if (disabled || !selected || selected === square) return
      tryMove(selected, square)
      setSelected(null)
    },
    [disabled, selected, tryMove],
  )

  const finishPromotion = useCallback(
    (piece: string) => {
      if (!pendingPromotion) return
      onMove(pendingPromotion.from, pendingPromotion.to, piece)
      setPendingPromotion(null)
      setSelected(null)
    },
    [pendingPromotion, onMove],
  )

  const rows = orientation === 'white' ? ORDER : [...ORDER].reverse()
  const cols = orientation === 'white' ? ORDER : [...ORDER].reverse()

  return (
    <div className="chess-board">
      {rows.map((ri, displayRow) =>
        cols.map((ci, displayCol) => {
          const cell = board[ri][ci]
          const square = FILES[ci] + (8 - ri)
          const isLight = (ri + ci) % 2 === 0
          const pieceKey = cell ? cell.color + cell.type : ''
          const hasOwnPiece = cell?.color === turn
          const flags = legalTargets.get(square)
          const isLegal = flags !== undefined
          const isCapture = isLegal && /[ec]/.test(flags!) // capture / en-passant
          const isLastMove =
            lastMove?.from === square || lastMove?.to === square

          return (
            <div
              key={square}
              className={[
                'square',
                isLight ? 'light' : 'dark',
                square === selected ? 'selected' : '',
                isLastMove ? 'highlight' : '',
                square === checkSquare ? 'check' : '',
              ]
                .filter(Boolean)
                .join(' ')}
              onClick={() => handleSquareClick(square, !!hasOwnPiece)}
              onDragOver={(e) => e.preventDefault()}
              onDrop={() => handleDrop(square)}
            >
              {displayCol === 0 && (
                <span className="square-coord rank">{8 - ri}</span>
              )}
              {displayRow === 7 && (
                <span className="square-coord file">{FILES[ci]}</span>
              )}

              {isLegal && !isCapture && <div className="legal-dot" />}
              {isLegal && isCapture && <div className="legal-ring" />}

              <AnimatePresence mode="wait">
                {cell && (
                  <motion.img
                    key={`${square}-${pieceKey}`}
                    src={pieceSrc(pieceKey)}
                    alt={pieceKey}
                    className="chess-piece select-none"
                    draggable={!disabled && hasOwnPiece}
                    onDragStart={() => !disabled && hasOwnPiece && setSelected(square)}
                    initial={{ scale: 0.85, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    exit={{ scale: 0.85, opacity: 0 }}
                    transition={{ type: 'spring', stiffness: 240, damping: 18 }}
                  />
                )}
              </AnimatePresence>
            </div>
          )
        }),
      )}

      {arrows.length > 0 && (
        <svg className="board-arrows" viewBox="0 0 8 8">
          {arrows.map((a, i) => {
            const s = squareToXY(a.from, orientation)
            const e = squareToXY(a.to, orientation)
            const dx = e.x - s.x
            const dy = e.y - s.y
            const len = Math.hypot(dx, dy) || 1
            const ux = dx / len
            const uy = dy / len
            const primary = a.kind !== 'secondary'
            const head = primary ? 0.36 : 0.28
            const width = primary ? 0.17 : 0.12
            const tipX = e.x - ux * 0.06
            const tipY = e.y - uy * 0.06
            const baseX = tipX - ux * head
            const baseY = tipY - uy * head
            const hw = head * 0.6 // half-width of the arrowhead
            return (
              <g
                key={`${a.from}${a.to}${i}`}
                className={primary ? 'arrow-primary' : 'arrow-secondary'}
              >
                <line
                  x1={s.x}
                  y1={s.y}
                  x2={baseX}
                  y2={baseY}
                  strokeWidth={width}
                  strokeLinecap="round"
                />
                <polygon
                  points={`${tipX},${tipY} ${baseX - uy * hw},${baseY + ux * hw} ${baseX + uy * hw},${baseY - ux * hw}`}
                />
              </g>
            )
          })}
        </svg>
      )}

      <AnimatePresence>
        {pendingPromotion && (
          <motion.div
            className="promo-overlay"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setPendingPromotion(null)}
          >
            <div className="promo-card" onClick={(e) => e.stopPropagation()}>
              {PROMO_PIECES.map((p) => (
                <button
                  key={p}
                  type="button"
                  className="promo-piece"
                  onClick={() => finishPromotion(p)}
                >
                  <img src={pieceSrc(turn + p)} alt={p} />
                </button>
              ))}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
})
