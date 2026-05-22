import { memo } from 'react'
import type { EngineLine } from '../hooks/useStockfish'
import { uciLineToSan, formatScore } from '../utils/chessFormat'

interface EngineLinesProps {
  fen: string
  lines: EngineLine[]
  depth: number
  thinking: boolean
  unavailable: boolean
}

export const EngineLines = memo(function EngineLines({
  fen,
  lines,
  depth,
  thinking,
  unavailable,
}: EngineLinesProps) {
  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs uppercase tracking-wide text-chess-text-dim">
          Engine
        </span>
        {!unavailable && (
          <span className="text-xs font-mono text-chess-text-dim">
            {thinking ? `analyzing · d${depth}` : `depth ${depth}`}
          </span>
        )}
      </div>

      {unavailable ? (
        <p className="text-sm text-chess-text-muted">
          In-browser engine unavailable — use “Deep analysis” for a server eval.
        </p>
      ) : lines.length === 0 ? (
        <p className="text-sm text-chess-text-muted">Starting engine…</p>
      ) : (
        <div className="space-y-1">
          {lines.map((line) => {
            const positive = line.mate !== 0 ? line.mate > 0 : line.scoreCp >= 0
            return (
              <div key={line.rank} className="flex gap-2 text-sm items-baseline">
                <span
                  className={`font-mono font-semibold w-14 shrink-0 ${
                    positive ? 'text-chess-text' : 'text-red-400'
                  }`}
                >
                  {formatScore(line.scoreCp, line.mate)}
                </span>
                <span className="font-mono text-chess-text-muted truncate">
                  {uciLineToSan(fen, line.pv).join(' ')}
                </span>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
})
