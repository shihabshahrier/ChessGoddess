import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { GameReview } from '../components/GameReview'
import { getSnapshotByToken } from '../api/snapshots'
import type { AnalysisSession, Move } from '../types'

export function SharePage() {
  const { snapshotId } = useParams() // route param holds the share token
  const [session, setSession] = useState<AnalysisSession | null>(null)
  const [moves, setMoves] = useState<Move[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!snapshotId) return
    let active = true
    getSnapshotByToken(snapshotId)
      .then(({ snapshot }) => {
        if (!active) return
        const data = snapshot.data as {
          session?: AnalysisSession
          moves?: Move[]
        }
        setSession(data.session ?? null)
        setMoves(data.moves ?? [])
        setLoading(false)
      })
      .catch((e) => {
        if (!active) return
        setError(e instanceof Error ? e.message : 'Snapshot not found')
        setLoading(false)
      })
    return () => {
      active = false
    }
  }, [snapshotId])

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-16 text-center">
        <h1 className="font-serif text-2xl text-chess-text mb-2">
          Snapshot Not Found
        </h1>
        <p className="text-chess-text-muted">{error}</p>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-20 text-center">
        <div className="text-5xl mb-4 animate-pulse">♟️</div>
        <p className="text-chess-text-muted">Loading shared analysis…</p>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="text-center mb-8">
        <h1 className="font-serif text-2xl font-semibold text-chess-text mb-1">
          Shared Analysis
        </h1>
        <p className="text-chess-text-muted">
          Read-only snapshot · {moves.length} moves
        </p>
      </div>
      {moves.length === 0 ? (
        <p className="text-chess-text-muted text-center">
          This snapshot has no analyzed moves.
        </p>
      ) : (
        <GameReview
          moves={moves}
          accuracyWhite={session?.accuracy_white ?? 0}
          accuracyBlack={session?.accuracy_black ?? 0}
        />
      )}
    </div>
  )
}
