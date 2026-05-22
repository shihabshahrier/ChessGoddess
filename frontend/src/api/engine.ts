// Server-side engine evaluation — deeper, on-demand analysis.
import { apiClient } from './client'

export interface ServerLine {
  rank: number
  depth: number
  score_cp: number // centipawns, side-to-move POV
  mate: number // moves to mate, side-to-move POV, 0 if none
  pv: string[] // UCI moves
}

export interface ServerEvaluation {
  fen: string
  depth: number
  best_move: string
  lines: ServerLine[]
}

/** Evaluate a position on the server engine (deeper than the in-browser one). */
export async function evaluatePosition(
  fen: string,
  depth = 18,
  multipv = 3,
): Promise<ServerEvaluation> {
  const { data } = await apiClient.post('/engine/evaluate', {
    fen,
    depth,
    multipv,
  })
  return data
}
