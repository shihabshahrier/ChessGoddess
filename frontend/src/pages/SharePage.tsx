import { useState, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { ChessBoard } from '../components/ChessBoard'
import { EvalBar } from '../components/EvalBar'
import { MoveList } from '../components/MoveList'
import { Timeline } from '../components/Timeline'

interface Snapshot {
  id: string
  session_id: string
  share_token: string
  data: {
    session: {
      id: string
      depth: number
      status: string
      created_at: string
    }
    moves: Move[]
  }
  created_at: string
}

interface Move {
  id: string
  move_number: number
  san: string
  classification?: 'blunder' | 'mistake' | 'inaccuracy' | 'good' | 'excellent' | 'best'
  evaluation?: number
  fen?: string
  best_move?: string
}

const mockSnapshot: Snapshot = {
  id: 'snap_123',
  session_id: 'sess_456',
  share_token: 'abc123',
  created_at: '2026-05-19T10:00:00Z',
  data: {
    session: { id: 'sess_456', depth: 20, status: 'completed', created_at: '2026-05-19T10:00:00Z' },
    moves: [
      { id: '1', move_number: 1, san: 'e4', classification: 'best', evaluation: 0.3, fen: 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1' },
      { id: '2', move_number: 2, san: 'e5', classification: 'best', evaluation: 0.2, fen: 'rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2' },
      { id: '3', move_number: 3, san: 'Nf3', classification: 'best', evaluation: 0.4, fen: 'rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2' },
      { id: '4', move_number: 4, san: 'Nc6', classification: 'good', evaluation: 0.3, fen: 'r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3' },
      { id: '5', move_number: 5, san: 'Bb5', classification: 'best', evaluation: 0.5, fen: 'r1bqkbnr/pppp1ppp/2n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 3 3' },
      { id: '6', move_number: 6, san: 'a6', classification: 'inaccuracy', evaluation: -0.2, fen: 'r1bqkbnr/1ppp1ppp/p1n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4' },
      { id: '7', move_number: 7, san: 'Bxc6', classification: 'good', evaluation: 0.3, fen: 'r1bqkbnr/1ppp1ppp/p1B5/4p3/4P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 0 4' },
      { id: '8', move_number: 8, san: 'dxc6', classification: 'mistake', evaluation: -0.8, fen: 'r1bqkbnr/1pp2ppp/p1p5/4p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 5' },
    ],
  },
}

export function SharePage() {
  const { snapshotId } = useParams()
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null)
  const [loading, setLoading] = useState(true)
  const [currentMoveIndex, setCurrentMoveIndex] = useState(-1)

  useEffect(() => {
    // TODO: Fetch from API
    // fetch(`/api/v1/snapshots/${snapshotId}`)
    setTimeout(() => {
      setSnapshot(mockSnapshot)
      setLoading(false)
    }, 500)
  }, [snapshotId])

  const moves = snapshot?.data.moves || []
  const currentMove = currentMoveIndex >= 0 ? moves[currentMoveIndex] : null
  const currentFEN = currentMove?.fen || 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'
  const currentEval = currentMove?.evaluation || 0
  const isBlunder = currentMove?.classification === 'blunder'

  const handleMoveSelect = useCallback((index: number) => {
    setCurrentMoveIndex(index)
  }, [])

  const handleMoveClick = useCallback((moveId: string) => {
    const index = moves.findIndex(m => m.id === moveId)
    setCurrentMoveIndex(index)
  }, [moves])

  const handleNext = useCallback(() => {
    setCurrentMoveIndex(prev => Math.min(prev + 1, moves.length - 1))
  }, [moves.length])

  const handlePrev = useCallback(() => {
    setCurrentMoveIndex(prev => Math.max(prev - 1, -1))
  }, [])

  const handleFirst = useCallback(() => {
    setCurrentMoveIndex(-1)
  }, [])

  const handleLast = useCallback(() => {
    setCurrentMoveIndex(moves.length - 1)
  }, [moves.length])

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 text-center">
        <div className="animate-pulse">
          <div className="h-8 bg-chess-surface rounded w-48 mx-auto mb-4" />
          <div className="h-4 bg-chess-surface rounded w-64 mx-auto" />
        </div>
      </div>
    )
  }

  if (!snapshot) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 text-center">
        <h1 className="font-serif text-2xl text-chess-text mb-4">Snapshot Not Found</h1>
        <p className="text-chess-text-muted">This analysis may have been removed.</p>
      </div>
    )
  }

  const blunderCount = moves.filter(m => m.classification === 'blunder').length
  const mistakeCount = moves.filter(m => m.classification === 'mistake').length

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="text-center mb-8">
        <h1 className="font-serif text-2xl font-semibold text-chess-text mb-2">
          Shared Analysis
        </h1>
        <p className="text-chess-text-muted">
          Immutable snapshot • {moves.length} moves • {blunderCount} blunders • {mistakeCount} mistakes
        </p>
        <p className="text-chess-text-dim text-sm mt-1 font-mono">
          {snapshot.share_token}
        </p>
      </div>

      <div className="grid lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          <div className="flex gap-4">
            <EvalBar evaluation={currentEval} />
            <ChessBoard 
              fen={currentFEN}
              interactive={true}
              isBlunder={isBlunder}
            />
          </div>
          
          <div className="mt-6 flex items-center justify-between bg-chess-surface border border-chess-border rounded-xl p-4">
            <div className="flex items-center gap-2">
              <button onClick={handleFirst} className="p-2 rounded-lg hover:bg-chess-elevated transition-colors">⏮</button>
              <button onClick={handlePrev} className="p-2 rounded-lg hover:bg-chess-elevated transition-colors">◀</button>
              <button onClick={handleNext} className="p-2 rounded-lg hover:bg-chess-elevated transition-colors">▶</button>
              <button onClick={handleLast} className="p-2 rounded-lg hover:bg-chess-elevated transition-colors">⏭</button>
            </div>
            <div className="text-chess-text-muted font-mono">
              Move {currentMoveIndex + 1} / {moves.length}
            </div>
          </div>
          
          <AnimatePresence mode="wait">
            {currentMove && (
              <motion.div 
                key={currentMove.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                transition={{ type: 'spring', stiffness: 200, damping: 20 }}
                className="mt-6 bg-chess-surface border border-chess-border rounded-xl p-6"
              >
                <div className="flex items-center gap-3 mb-3">
                  <h3 className="font-serif text-lg font-semibold text-chess-gold">
                    Move {currentMove.move_number}: {currentMove.san}
                  </h3>
                  {currentMove.classification && (
                    <span className={`px-2 py-1 rounded text-xs font-medium ${
                      currentMove.classification === 'blunder' ? 'bg-red-500/20 text-red-400' :
                      currentMove.classification === 'mistake' ? 'bg-orange-500/20 text-orange-400' :
                      currentMove.classification === 'inaccuracy' ? 'bg-yellow-500/20 text-yellow-400' :
                      currentMove.classification === 'good' ? 'bg-green-500/20 text-green-400' :
                      'bg-chess-gold/20 text-chess-gold'
                    }`}>
                      {currentMove.classification}
                    </span>
                  )}
                </div>
                <p className="text-chess-text-muted">
                  Evaluation: {currentEval > 0 ? '+' : ''}{currentEval.toFixed(2)}
                </p>
              </motion.div>
            )}
          </AnimatePresence>
          
          <div className="mt-6 bg-chess-surface border border-chess-border rounded-xl p-4">
            <Timeline 
              moves={moves} 
              currentMoveIndex={currentMoveIndex} 
              onMoveSelect={handleMoveSelect} 
            />
          </div>
        </div>
        
        <div className="bg-chess-surface border border-chess-border rounded-xl p-4">
          <h2 className="font-serif text-lg font-semibold text-chess-text mb-4">Moves</h2>
          <MoveList 
            moves={moves} 
            interactive={true} 
            onMoveClick={handleMoveClick} 
            activeMoveId={currentMove?.id} 
          />
          
          <div className="mt-6 pt-6 border-t border-chess-border">
            <div className="text-center text-chess-text-dim text-sm mb-4">
              Read-only shared analysis
            </div>
            <button 
              onClick={() => navigator.clipboard.writeText(window.location.href)}
              className="w-full border border-chess-gold text-chess-gold py-2 rounded-lg font-medium hover:bg-chess-gold/10 transition-colors"
            >
              Copy Share Link
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
