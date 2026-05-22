import { Chess } from 'chess.js'

/** Convert a UCI principal variation into SAN, replayed from `fen`. */
export function uciLineToSan(
  fen: string,
  uciMoves: string[],
  limit = 8,
): string[] {
  const sans: string[] = []
  try {
    const game = new Chess(fen)
    for (const uci of uciMoves.slice(0, limit)) {
      const move = game.move({
        from: uci.slice(0, 2),
        to: uci.slice(2, 4),
        promotion: uci.length > 4 ? uci[4] : undefined,
      })
      sans.push(move.san)
    }
  } catch {
    /* malformed PV — return what was parsed so far */
  }
  return sans
}

/** Format a White-POV score for display: "+1.40", "-0.30", "M5", "-M2". */
export function formatScore(scoreCp: number, mate: number): string {
  if (mate !== 0) return `${mate > 0 ? '' : '-'}M${Math.abs(mate)}`
  const pawns = scoreCp / 100
  return `${pawns > 0 ? '+' : ''}${pawns.toFixed(2)}`
}
