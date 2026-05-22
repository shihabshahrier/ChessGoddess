import { useState, useCallback, useMemo, useEffect } from 'react'
import { Chess } from 'chess.js'
import { motion, AnimatePresence } from 'framer-motion'
import { InteractiveBoard } from './InteractiveBoard'
import type { BoardArrow } from './InteractiveBoard'
import { EvalBar } from './EvalBar'
import { EvalGraph } from './EvalGraph'
import { MoveList } from './MoveList'
import { explainMove } from '../api/ai'
import type { Move } from '../types'
import { CLASSIFICATION_META } from '../utils/classification'
import { uciLineToSan } from '../utils/chessFormat'

interface GameReviewProps {
  moves: Move[]
  accuracyWhite: number
  accuracyBlack: number
  /** When set, per-move AI explanations are available (authenticated). */
  sessionId?: string
}

const noLegalMoves = () => []
const noMove = () => {}

export function GameReview({
  moves,
  accuracyWhite,
  accuracyBlack,
  sessionId,
}: GameReviewProps) {
  const [currentIndex, setCurrentIndex] = useState(0)
  const [orientation, setOrientation] = useState<'white' | 'black'>('white')
  const [explanations, setExplanations] = useState<Record<string, string>>({})
  const [explaining, setExplaining] = useState(false)
  const [explainError, setExplainError] = useState('')

  // Replay SANs into positions: positions[k] is the board after k moves.
  const positions = useMemo(() => {
    const list: { fen: string; from?: string; to?: string }[] = []
    const game = new Chess()
    list.push({ fen: game.fen() })
    for (const m of moves) {
      try {
        const mv = game.move(m.san)
        list.push({ fen: game.fen(), from: mv.from, to: mv.to })
      } catch {
        break
      }
    }
    return list
  }, [moves])

  const total = positions.length - 1
  const idx = Math.min(currentIndex, total)
  const pos = positions[idx] ?? positions[0]
  const playedMove = idx > 0 ? moves[idx - 1] : null // move that reached this position
  const nextBest = moves[idx] // engine's pick for the current position

  const evaluations = useMemo(() => moves.map((m) => m.evaluation), [moves])

  const lastMove =
    pos.from && pos.to ? { from: pos.from, to: pos.to } : null

  const arrows = useMemo<BoardArrow[]>(() => {
    const uci = nextBest?.best_move
    if (!uci || uci.length < 4) return []
    return [{ from: uci.slice(0, 2), to: uci.slice(2, 4), kind: 'primary' }]
  }, [nextBest])

  const goto = useCallback(
    (i: number) => setCurrentIndex(Math.max(0, Math.min(total, i))),
    [total],
  )

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'ArrowLeft') goto(idx - 1)
      else if (e.key === 'ArrowRight') goto(idx + 1)
      else if (e.key === 'Home') goto(0)
      else if (e.key === 'End') goto(total)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [idx, total, goto])

  const handleExplain = useCallback(async () => {
    if (!sessionId || !playedMove || explanations[playedMove.id]) return
    setExplaining(true)
    setExplainError('')
    try {
      const { explanation } = await explainMove(playedMove.id, sessionId)
      setExplanations((prev) => ({ ...prev, [playedMove.id]: explanation }))
    } catch (e) {
      setExplainError(e instanceof Error ? e.message : 'Explanation failed')
    } finally {
      setExplaining(false)
    }
  }, [sessionId, playedMove, explanations])

  const moveListItems = useMemo(
    () =>
      moves.map((m) => ({
        id: m.id,
        san: m.san,
        classification: m.classification,
      })),
    [moves],
  )

  const handleMoveClick = useCallback(
    (moveId: string) => {
      const i = moves.findIndex((m) => m.id === moveId)
      if (i >= 0) goto(i + 1)
    },
    [moves, goto],
  )

  const summary = useMemo(() => {
    const c: Partial<Record<string, number>> = {}
    for (const m of moves) c[m.classification] = (c[m.classification] ?? 0) + 1
    return c
  }, [moves])

  const meta = playedMove
    ? CLASSIFICATION_META[playedMove.classification]
    : null
  const explanation = playedMove ? explanations[playedMove.id] : undefined

  return (
    <div className="grid lg:grid-cols-3 gap-8">
      <div className="lg:col-span-2">
        <div className="flex gap-3 items-stretch">
          <EvalBar evaluation={playedMove ? playedMove.evaluation : 0} />
          <div className="flex-1 min-w-0">
            <InteractiveBoard
              fen={pos.fen}
              onMove={noMove}
              getLegalMoves={noLegalMoves}
              orientation={orientation}
              lastMove={lastMove}
              arrows={arrows}
              disabled
            />
          </div>
        </div>

        <div className="mt-4">
          <EvalGraph
            evaluations={evaluations}
            currentIndex={idx}
            onSelect={goto}
          />
        </div>

        <div className="mt-4 flex items-center justify-between bg-chess-surface border border-chess-border rounded-xl p-3">
          <div className="flex items-center gap-1">
            <button
              onClick={() => goto(0)}
              className="p-2 rounded-lg hover:bg-chess-elevated transition-colors"
            >
              ⏮
            </button>
            <button
              onClick={() => goto(idx - 1)}
              className="p-2 rounded-lg hover:bg-chess-elevated transition-colors"
            >
              ◀
            </button>
            <button
              onClick={() => goto(idx + 1)}
              className="p-2 rounded-lg hover:bg-chess-elevated transition-colors"
            >
              ▶
            </button>
            <button
              onClick={() => goto(total)}
              className="p-2 rounded-lg hover:bg-chess-elevated transition-colors"
            >
              ⏭
            </button>
            <button
              onClick={() =>
                setOrientation((o) => (o === 'white' ? 'black' : 'white'))
              }
              className="p-2 rounded-lg hover:bg-chess-elevated transition-colors"
            >
              ⇅
            </button>
          </div>
          <span className="text-chess-text-muted font-mono text-sm">
            {idx} / {total}
          </span>
        </div>

        <AnimatePresence mode="wait">
          {playedMove && meta && (
            <motion.div
              key={playedMove.id}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -12 }}
              transition={{ type: 'spring', stiffness: 220, damping: 22 }}
              className="mt-4 bg-chess-surface border border-chess-border rounded-xl p-5"
            >
              <div className="flex items-center gap-3 flex-wrap">
                <h3 className="font-serif text-lg font-semibold text-chess-text">
                  {Math.ceil(idx / 2)}
                  {idx % 2 === 1 ? '.' : '...'} {playedMove.san}
                </h3>
                <span
                  className={`px-2 py-0.5 rounded text-xs font-medium ${meta.badge}`}
                >
                  {meta.icon} {meta.label}
                </span>
                {playedMove.cp_loss > 5 && (
                  <span className="text-chess-text-muted text-sm">
                    lost {(playedMove.cp_loss / 100).toFixed(2)}
                  </span>
                )}
                <span className="text-chess-text-dim text-sm font-mono ml-auto">
                  {playedMove.accuracy.toFixed(0)}% accurate
                </span>
              </div>

              {playedMove.classification !== 'best' &&
                playedMove.best_line && (
                  <p className="text-chess-text-muted text-sm mt-2 font-mono">
                    Best:{' '}
                    {uciLineToSan(
                      positions[idx - 1]?.fen ?? pos.fen,
                      playedMove.best_line.split(' '),
                      5,
                    ).join(' ')}
                  </p>
                )}

              {sessionId && (
                <div className="mt-3">
                  {explanation ? (
                    <p className="text-chess-text text-sm leading-relaxed border-l-2 border-chess-gold pl-3">
                      {explanation}
                    </p>
                  ) : (
                    <button
                      onClick={handleExplain}
                      disabled={explaining}
                      className="text-sm text-chess-gold hover:text-chess-gold-light transition-colors disabled:opacity-50"
                    >
                      {explaining ? 'Thinking…' : '🤖 Explain this move'}
                    </button>
                  )}
                  {explainError && (
                    <p className="text-red-400 text-xs mt-1">{explainError}</p>
                  )}
                </div>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      <div className="space-y-6">
        <div className="bg-chess-surface border border-chess-border rounded-xl p-4">
          <h2 className="font-serif text-lg font-semibold text-chess-text mb-3">
            Accuracy
          </h2>
          <div className="flex gap-4">
            {(
              [
                ['White', accuracyWhite],
                ['Black', accuracyBlack],
              ] as const
            ).map(([side, acc]) => (
              <div
                key={side}
                className="flex-1 text-center bg-chess-elevated rounded-lg py-3"
              >
                <div className="text-2xl font-bold text-chess-gold">
                  {acc.toFixed(1)}
                </div>
                <div className="text-xs text-chess-text-muted">{side}</div>
              </div>
            ))}
          </div>
          <div className="flex flex-wrap gap-x-3 gap-y-1 mt-3 text-xs">
            {(
              ['blunder', 'mistake', 'inaccuracy', 'brilliant'] as const
            ).map((k) =>
              summary[k] ? (
                <span key={k} className={CLASSIFICATION_META[k].color}>
                  {summary[k]} {CLASSIFICATION_META[k].label.toLowerCase()}
                </span>
              ) : null,
            )}
          </div>
        </div>

        <div className="bg-chess-surface border border-chess-border rounded-xl p-4">
          <h2 className="font-serif text-lg font-semibold text-chess-text mb-3">
            Moves
          </h2>
          <MoveList
            moves={moveListItems}
            activeMoveId={playedMove?.id}
            onMoveClick={handleMoveClick}
          />
        </div>
      </div>
    </div>
  )
}
