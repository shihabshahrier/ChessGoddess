import { useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { ChessBoard } from '../components/ChessBoard'
import { EvalBar } from '../components/EvalBar'
import { MoveList } from '../components/MoveList'
import { Timeline } from '../components/Timeline'

interface Move {
  id: string
  moveNumber: number
  san: string
  classification?: 'blunder' | 'mistake' | 'inaccuracy' | 'good' | 'excellent' | 'best'
  evaluation?: number
  fen?: string
}

const mockMoves: Move[] = [
  { id: '1', moveNumber: 1, san: 'e4', classification: 'best', evaluation: 0.3, fen: 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1' },
  { id: '2', moveNumber: 2, san: 'e5', classification: 'best', evaluation: 0.2, fen: 'rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2' },
  { id: '3', moveNumber: 3, san: 'Nf3', classification: 'best', evaluation: 0.4, fen: 'rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2' },
  { id: '4', moveNumber: 4, san: 'Nc6', classification: 'good', evaluation: 0.3, fen: 'r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3' },
  { id: '5', moveNumber: 5, san: 'Bb5', classification: 'best', evaluation: 0.5, fen: 'r1bqkbnr/pppp1ppp/2n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 3 3' },
  { id: '6', moveNumber: 6, san: 'a6', classification: 'inaccuracy', evaluation: -0.2, fen: 'r1bqkbnr/1ppp1ppp/p1n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4' },
  { id: '7', moveNumber: 7, san: 'Bxc6', classification: 'good', evaluation: 0.3, fen: 'r1bqkbnr/1ppp1ppp/p1B5/4p3/4P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 0 4' },
  { id: '8', moveNumber: 8, san: 'dxc6', classification: 'mistake', evaluation: -0.8, fen: 'r1bqkbnr/1pp2ppp/p1p5/4p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 5' },
]

export function ReviewPage() {
  const [currentMoveIndex, setCurrentMoveIndex] = useState(-1)

  const currentMove = currentMoveIndex >= 0 ? mockMoves[currentMoveIndex] : null
  const currentFEN = currentMove?.fen || 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1'
  const currentEval = currentMove?.evaluation || 0
  const isBlunder = currentMove?.classification === 'blunder'

  const handleMoveSelect = useCallback((index: number) => {
    setCurrentMoveIndex(index)
  }, [])

  const handleMoveClick = useCallback((moveId: string) => {
    const index = mockMoves.findIndex(m => m.id === moveId)
    setCurrentMoveIndex(index)
  }, [])

  const handleNext = useCallback(() => {
    setCurrentMoveIndex(prev => Math.min(prev + 1, mockMoves.length - 1))
  }, [])

  const handlePrev = useCallback(() => {
    setCurrentMoveIndex(prev => Math.max(prev - 1, -1))
  }, [])

  const handleFirst = useCallback(() => {
    setCurrentMoveIndex(-1)
  }, [])

  const handleLast = useCallback(() => {
    setCurrentMoveIndex(mockMoves.length - 1)
  }, [])

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
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
              <button 
                onClick={handleFirst}
                className="p-2 rounded-lg hover:bg-chess-elevated transition-colors"
              >
                ⏮
              </button>
              <button 
                onClick={handlePrev}
                className="p-2 rounded-lg hover:bg-chess-elevated transition-colors"
              >
                ◀
              </button>
              <button 
                onClick={handleNext}
                className="p-2 rounded-lg hover:bg-chess-elevated transition-colors"
              >
                ▶
              </button>
              <button 
                onClick={handleLast}
                className="p-2 rounded-lg hover:bg-chess-elevated transition-colors"
              >
                ⏭
              </button>
            </div>
            
            <div className="text-chess-text-muted font-mono">
              Move {currentMoveIndex + 1} / {mockMoves.length}
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
                    Move {currentMove.moveNumber}: {currentMove.san}
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
              moves={mockMoves} 
              currentMoveIndex={currentMoveIndex} 
              onMoveSelect={handleMoveSelect} 
            />
          </div>
        </div>
        
        <div className="bg-chess-surface border border-chess-border rounded-xl p-4">
          <h2 className="font-serif text-lg font-semibold text-chess-text mb-4">Review Timeline</h2>
          <MoveList 
            moves={mockMoves} 
            interactive={true} 
            onMoveClick={handleMoveClick} 
            activeMoveId={currentMove?.id} 
          />
          
          <div className="mt-6 pt-6 border-t border-chess-border">
            <button className="w-full bg-chess-gold text-chess-bg py-2 rounded-lg font-medium hover:bg-chess-gold-light transition-colors">
              Share Analysis
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
