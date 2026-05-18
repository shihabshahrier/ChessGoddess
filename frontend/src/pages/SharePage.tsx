import { useParams } from 'react-router-dom'
import { ChessBoard } from '../components/ChessBoard'
import { EvalBar } from '../components/EvalBar'
import { MoveList } from '../components/MoveList'

export function SharePage() {
  const { snapshotId } = useParams()

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="text-center mb-8">
        <h1 className="font-serif text-2xl font-semibold text-chess-text mb-2">
          Shared Analysis
        </h1>
        <p className="text-chess-text-muted">
          Viewing immutable snapshot: <span className="font-mono text-chess-gold">{snapshotId}</span>
        </p>
      </div>

      <div className="grid lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          <div className="flex gap-4">
            <EvalBar evaluation={0.5} />
            <ChessBoard 
              fen="rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
              interactive={true}
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
