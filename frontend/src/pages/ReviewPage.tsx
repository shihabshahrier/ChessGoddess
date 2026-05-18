import { useParams } from 'react-router-dom'
import { ChessBoard } from '../components/ChessBoard'
import { EvalBar } from '../components/EvalBar'
import { MoveList } from '../components/MoveList'

export function ReviewPage() {
  const { id } = useParams()

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="grid lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          <div className="flex gap-4">
            <EvalBar evaluation={0.5} />
            <ChessBoard 
              fen="rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
              interactive={true}
            />
          </div>
          
          <div className="mt-6 bg-chess-surface border border-chess-border rounded-xl p-6">
            <h3 className="font-serif text-lg font-semibold text-chess-gold mb-2">AI Insight</h3>
            <p className="text-chess-text-muted">
              Analysis insights will appear here after AI processing.
            </p>
          </div>
        </div>
        
        <div className="bg-chess-surface border border-chess-border rounded-xl p-4">
          <h2 className="font-serif text-lg font-semibold text-chess-text mb-4">Review Timeline</h2>
          <MoveList moves={[]} interactive={true} />
          
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
