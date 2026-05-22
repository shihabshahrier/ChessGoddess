import { useState, useCallback, useMemo, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { InteractiveBoard } from '../components/InteractiveBoard'
import type { BoardArrow } from '../components/InteractiveBoard'
import { EvalBar } from '../components/EvalBar'
import { EngineLines } from '../components/EngineLines'
import { useChessGame } from '../hooks/useChessGame'
import { useStockfish } from '../hooks/useStockfish'
import { evaluatePosition } from '../api/engine'
import type { ServerEvaluation } from '../api/engine'
import { uciLineToSan, formatScore } from '../utils/chessFormat'

export function AnalysisPage() {
  const location = useLocation()
  // A FEN may arrive from the board scanner (router state) or a shareable ?fen= link.
  const initialFen =
    (location.state as { fen?: string } | null)?.fen ??
    new URLSearchParams(location.search).get('fen') ??
    undefined

  const game = useChessGame(initialFen)
  const engine = useStockfish(18)
  const { analyze, stop } = engine

  const [orientation, setOrientation] = useState<'white' | 'black'>('white')
  const [copied, setCopied] = useState(false)
  const [serverEval, setServerEval] = useState<ServerEvaluation | null>(null)
  const [serverLoading, setServerLoading] = useState(false)
  const [serverError, setServerError] = useState('')

  // Re-run the in-browser engine whenever the position changes.
  useEffect(() => {
    if (!engine.unavailable) analyze(game.fen)
  }, [game.fen, analyze, engine.unavailable])

  useEffect(() => () => stop(), [stop])

  const lastMove = useMemo(() => {
    const last = game.history[game.history.length - 1]
    return last ? { from: last.from, to: last.to } : null
  }, [game.history])

  // Engine suggestions → board arrows (best is bold, alternatives faint).
  const arrows = useMemo<BoardArrow[]>(
    () =>
      engine.lines
        .filter((l) => l.pv.length > 0)
        .slice(0, 3)
        .map((l, idx) => ({
          from: l.pv[0].slice(0, 2),
          to: l.pv[0].slice(2, 4),
          kind: idx === 0 ? 'primary' : 'secondary',
        })),
    [engine.lines],
  )

  const best = engine.lines[0]
  const evalPawns = best ? best.scoreCp / 100 : 0
  const evalMate = best ? best.mate : 0

  const handleMove = useCallback(
    (from: string, to: string, promotion?: string) => {
      game.makeMove(from, to, promotion)
      setServerEval(null)
    },
    [game],
  )

  const handleFlip = useCallback(
    () => setOrientation((o) => (o === 'white' ? 'black' : 'white')),
    [],
  )

  const handleCopyFen = useCallback(() => {
    navigator.clipboard.writeText(game.fen)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }, [game.fen])

  const handleNewGame = useCallback(() => {
    game.reset()
    setServerEval(null)
    setServerError('')
  }, [game])

  const runDeepAnalysis = useCallback(async () => {
    setServerLoading(true)
    setServerError('')
    try {
      setServerEval(await evaluatePosition(game.fen, 20, 3))
    } catch (e) {
      setServerError(e instanceof Error ? e.message : 'Server analysis failed')
    } finally {
      setServerLoading(false)
    }
  }, [game.fen])

  const moveRows = useMemo(() => {
    const rows: { no: number; white?: string; black?: string }[] = []
    for (let i = 0; i < game.history.length; i += 2) {
      rows.push({
        no: i / 2 + 1,
        white: game.history[i]?.san,
        black: game.history[i + 1]?.san,
      })
    }
    return rows
  }, [game.history])

  const status = game.isCheckmate
    ? `Checkmate — ${game.turn === 'w' ? 'Black' : 'White'} wins`
    : game.isStalemate
      ? 'Stalemate — draw'
      : game.isDraw
        ? 'Draw'
        : game.inCheck
          ? `Check — ${game.turn === 'w' ? 'White' : 'Black'} to move`
          : `${game.turn === 'w' ? 'White' : 'Black'} to move`

  // The server reports side-to-move POV; flip to White POV for display.
  const serverSign = serverEval?.fen.split(' ')[1] === 'b' ? -1 : 1

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <h1 className="font-serif text-3xl font-bold text-chess-text mb-2">
        Analysis Board
      </h1>
      <p className="text-chess-text-muted mb-8">
        Click or drag pieces to explore. The engine evaluates every position
        live and points the way.
      </p>

      <div className="grid lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          <div className="flex gap-3 items-stretch">
            <EvalBar evaluation={evalPawns} mate={evalMate} />
            <div className="flex-1 min-w-0">
              <InteractiveBoard
                fen={game.fen}
                onMove={handleMove}
                getLegalMoves={game.getLegalMoves}
                orientation={orientation}
                lastMove={lastMove}
                arrows={arrows}
                disabled={game.isGameOver}
              />
            </div>
          </div>

          <div className="mt-6 flex flex-wrap items-center gap-3 bg-chess-surface border border-chess-border rounded-xl p-4">
            <button
              onClick={game.undo}
              disabled={game.history.length === 0}
              className="px-4 py-2 rounded-lg bg-chess-elevated text-chess-text hover:bg-chess-border transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >
              ↩ Undo
            </button>
            <button
              onClick={handleFlip}
              className="px-4 py-2 rounded-lg bg-chess-elevated text-chess-text hover:bg-chess-border transition-colors"
            >
              ⇅ Flip board
            </button>
            <button
              onClick={handleNewGame}
              className="px-4 py-2 rounded-lg bg-chess-elevated text-chess-text hover:bg-chess-border transition-colors"
            >
              ⟲ New game
            </button>
            <button
              onClick={handleCopyFen}
              className="px-4 py-2 rounded-lg bg-chess-elevated text-chess-text hover:bg-chess-border transition-colors"
            >
              {copied ? '✓ Copied' : '⎘ Copy FEN'}
            </button>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-chess-surface border border-chess-border rounded-xl p-4">
            <div className="flex items-center justify-between mb-3">
              <h2 className="font-serif text-lg font-semibold text-chess-text">
                Position
              </h2>
              <span
                className={`text-sm font-medium ${
                  game.isGameOver
                    ? 'text-chess-gold'
                    : game.inCheck
                      ? 'text-red-400'
                      : 'text-chess-text-muted'
                }`}
              >
                {status}
              </span>
            </div>
            <EngineLines
              fen={game.fen}
              lines={engine.lines}
              depth={engine.depth}
              thinking={engine.thinking}
              unavailable={engine.unavailable}
            />

            <div className="mt-4 pt-4 border-t border-chess-border">
              <button
                onClick={runDeepAnalysis}
                disabled={serverLoading}
                className="w-full px-4 py-2 rounded-lg bg-chess-gold text-chess-bg font-medium hover:bg-chess-gold-light transition-colors disabled:opacity-50"
              >
                {serverLoading ? 'Analyzing on server…' : '⌁ Deep analysis (server)'}
              </button>
              {serverError && (
                <p className="text-red-400 text-xs mt-2">{serverError}</p>
              )}
              {serverEval && (
                <div className="mt-3 space-y-1">
                  <p className="text-xs text-chess-text-dim">
                    Server · depth {serverEval.depth}
                  </p>
                  {serverEval.lines.slice(0, 3).map((l) => (
                    <div
                      key={l.rank}
                      className="flex gap-2 text-sm items-baseline"
                    >
                      <span className="font-mono font-semibold w-14 shrink-0 text-chess-text">
                        {formatScore(l.score_cp * serverSign, l.mate * serverSign)}
                      </span>
                      <span className="font-mono text-chess-text-muted truncate">
                        {uciLineToSan(serverEval.fen, l.pv).join(' ')}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="bg-chess-surface border border-chess-border rounded-xl p-4">
            <h2 className="font-serif text-lg font-semibold text-chess-text mb-3">
              Moves
            </h2>
            {moveRows.length === 0 ? (
              <p className="text-chess-text-muted text-sm py-6 text-center">
                No moves yet — make the first move.
              </p>
            ) : (
              <div className="max-h-72 overflow-y-auto font-mono text-sm">
                {moveRows.map((row) => (
                  <div
                    key={row.no}
                    className="flex items-center gap-2 py-1 border-b border-chess-border/40 last:border-0"
                  >
                    <span className="w-8 text-chess-text-dim">{row.no}.</span>
                    <span className="w-20 text-chess-text">{row.white}</span>
                    <span className="w-20 text-chess-text">{row.black}</span>
                  </div>
                ))}
              </div>
            )}
            <div className="mt-4 pt-3 border-t border-chess-border">
              <p className="text-xs text-chess-text-dim mb-1">Current FEN</p>
              <p className="text-xs font-mono text-chess-text-muted break-all">
                {game.fen}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
