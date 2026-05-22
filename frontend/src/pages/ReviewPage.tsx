import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { GameReview } from '../components/GameReview'
import { getAnalysis, getAnalysisMoves } from '../api/games'
import type { AnalysisSession, Move } from '../types'

export function ReviewPage() {
  const { id } = useParams()
  const [session, setSession] = useState<AnalysisSession | null>(null)
  const [moves, setMoves] = useState<Move[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    let active = true
    let timer: number | undefined

    const load = async () => {
      try {
        const res = await getAnalysis(id)
        if (!active) return
        setSession(res.session)

        if (res.session.status === 'completed') {
          const m = await getAnalysisMoves(id)
          if (!active) return
          setMoves(m.moves ?? [])
          setLoading(false)
        } else if (res.session.status === 'failed') {
          setError('Analysis failed — try re-uploading the game.')
          setLoading(false)
        } else {
          // pending / running — poll until done
          timer = window.setTimeout(load, 2000)
        }
      } catch (e) {
        if (!active) return
        setError(e instanceof Error ? e.message : 'Could not load analysis')
        setLoading(false)
      }
    }

    load()
    return () => {
      active = false
      if (timer) clearTimeout(timer)
    }
  }, [id])

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-16 text-center">
        <h1 className="font-serif text-2xl text-chess-text mb-2">
          Analysis Unavailable
        </h1>
        <p className="text-chess-text-muted">{error}</p>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-20 text-center">
        <div className="text-5xl mb-4 animate-pulse">♟️</div>
        <h1 className="font-serif text-2xl text-chess-text mb-2">
          {session?.status === 'running'
            ? 'Analyzing your game…'
            : 'Queued for analysis…'}
        </h1>
        <p className="text-chess-text-muted">
          Stockfish is reviewing every move. This usually takes a moment.
        </p>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <h1 className="font-serif text-3xl font-bold text-chess-text mb-1">
        Game Review
      </h1>
      <p className="text-chess-text-muted mb-8">
        Step through every move — engine evaluation, classification, and AI
        insight.
      </p>
      {moves.length === 0 ? (
        <p className="text-chess-text-muted">
          No moves were analyzed for this game.
        </p>
      ) : (
        <GameReview
          moves={moves}
          accuracyWhite={session?.accuracy_white ?? 0}
          accuracyBlack={session?.accuracy_black ?? 0}
          sessionId={id}
        />
      )}
    </div>
  )
}
