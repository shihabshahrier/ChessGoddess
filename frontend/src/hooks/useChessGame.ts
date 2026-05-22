import { useRef, useState, useCallback } from 'react'
import { Chess } from 'chess.js'
import type { Move, Square } from 'chess.js'

export const START_FEN =
  'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'

export interface HistoryEntry {
  san: string
  fen: string
  from: string
  to: string
  color: 'w' | 'b'
}

export interface LegalMove {
  to: string
  flags: string
}

/** Construct a Chess instance, falling back to the start position on a bad FEN. */
function safeChess(fen: string): Chess {
  try {
    return new Chess(fen)
  } catch {
    return new Chess(START_FEN)
  }
}

/** Reports whether a FEN string is a legal, loadable chess position. */
export function isValidFen(fen: string): boolean {
  try {
    new Chess(fen)
    return true
  } catch {
    return false
  }
}

/**
 * Wraps a chess.js game with React state. The Chess instance lives in a ref
 * (mutable, authoritative); `fen` state drives re-renders.
 */
export function useChessGame(initialFen: string = START_FEN) {
  const gameRef = useRef(safeChess(initialFen))
  const [fen, setFen] = useState(() => gameRef.current.fen())
  const [history, setHistory] = useState<HistoryEntry[]>([])

  const makeMove = useCallback(
    (from: string, to: string, promotion: string = 'q'): boolean => {
      try {
        const mv = gameRef.current.move({ from, to, promotion })
        if (!mv) return false
        const next = gameRef.current.fen()
        setHistory((h) => [
          ...h,
          { san: mv.san, fen: next, from: mv.from, to: mv.to, color: mv.color },
        ])
        setFen(next)
        return true
      } catch {
        return false
      }
    },
    [],
  )

  const getLegalMoves = useCallback((square: string): LegalMove[] => {
    try {
      const moves = gameRef.current.moves({
        square: square as Square,
        verbose: true,
      }) as Move[]
      return moves.map((m) => ({ to: m.to, flags: m.flags }))
    } catch {
      return []
    }
  }, [])

  const undo = useCallback(() => {
    const undone = gameRef.current.undo()
    if (!undone) return
    setHistory((h) => h.slice(0, -1))
    setFen(gameRef.current.fen())
  }, [])

  const reset = useCallback((toFen: string = START_FEN) => {
    gameRef.current = safeChess(toFen)
    setHistory([])
    setFen(gameRef.current.fen())
  }, [])

  const game = gameRef.current

  return {
    fen,
    history,
    turn: game.turn() as 'w' | 'b',
    inCheck: game.inCheck(),
    isCheckmate: game.isCheckmate(),
    isStalemate: game.isStalemate(),
    isDraw: game.isDraw(),
    isGameOver: game.isGameOver(),
    makeMove,
    getLegalMoves,
    undo,
    reset,
  }
}
