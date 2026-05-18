import { useParams } from 'react-router-dom'
import { ChessBoard } from '../components/ChessBoard'
import { EvalBar } from '../components/EvalBar'
import { MoveList } from '../components/MoveList'

export function AnalysisPage() {
  const { id } = useParams()

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="grid lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          <div className="flex gap-4">
            <EvalBar evaluation={0.5} />
            <ChessBoard 
              fen="rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
              interactive={false}
            />
          </div>
        </div>
        
        <div className="bg-chess-surface border border-chess-border rounded-xl p-4">
          <h2 className="font-serif text-lg font-semibold text-chess-text mb-4">Moves</h2>
          <MoveList moves={[]} />
        </div>
      </div>
    </div>
  )
}
