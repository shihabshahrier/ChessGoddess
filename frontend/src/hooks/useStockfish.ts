import { useCallback, useEffect, useRef, useState } from 'react'

/** One engine principal variation, scored from White's POV. */
export interface EngineLine {
  rank: number
  depth: number
  scoreCp: number // white POV centipawns (0 when mate != 0)
  mate: number // white POV moves-to-mate, 0 if none
  pv: string[] // UCI long-algebraic moves
}

export interface EngineState {
  ready: boolean
  unavailable: boolean // worker failed to load — caller should degrade
  thinking: boolean
  depth: number
  lines: EngineLine[] // sorted by rank; lines[0] is best
}

const ENGINE_URL =
  (import.meta.env.VITE_STOCKFISH_URL as string) || '/stockfish/stockfish.js'
const MULTIPV = 3

function sideToMove(fen: string): 'w' | 'b' {
  return fen.split(' ')[1] === 'b' ? 'b' : 'w'
}

function parseInfo(text: string, side: 'w' | 'b'): EngineLine | null {
  const f = text.split(/\s+/)
  let depth = 0
  let rank = 1
  let scoreCp = 0
  let mate = 0
  let pv: string[] = []
  for (let i = 0; i < f.length; i++) {
    switch (f[i]) {
      case 'depth':
        depth = parseInt(f[i + 1]) || 0
        break
      case 'multipv':
        rank = parseInt(f[i + 1]) || 1
        break
      case 'score':
        if (f[i + 1] === 'cp') scoreCp = parseInt(f[i + 2]) || 0
        else if (f[i + 1] === 'mate') mate = parseInt(f[i + 2]) || 0
        break
      case 'pv':
        pv = f.slice(i + 1)
        i = f.length
        break
    }
  }
  if (pv.length === 0) return null
  // UCI scores are side-to-move POV — normalize to White's POV.
  const sign = side === 'w' ? 1 : -1
  return { rank, depth, scoreCp: scoreCp * sign, mate: mate * sign, pv }
}

/**
 * Runs Stockfish in a Web Worker for live, client-side evaluation.
 * One search at a time: a new analyze() stops the running search and queues
 * the latest position. Reports `unavailable` if the engine fails to load so
 * callers can fall back to the server engine.
 */
export function useStockfish(searchDepth = 14) {
  const [state, setState] = useState<EngineState>({
    ready: false,
    unavailable: false,
    thinking: false,
    depth: 0,
    lines: [],
  })

  const depthRef = useRef(searchDepth)
  depthRef.current = searchDepth

  const analyzeRef = useRef<(fen: string) => void>(() => {})
  const stopRef = useRef<() => void>(() => {})

  useEffect(() => {
    let worker: Worker
    try {
      worker = new Worker(ENGINE_URL)
    } catch {
      setState((s) => ({ ...s, unavailable: true }))
      return
    }

    let disposed = false
    let searching = false
    let pending: string | null = null
    let side: 'w' | 'b' = 'w'
    let rafScheduled = false
    const lines = new Map<number, EngineLine>()

    const send = (cmd: string) => worker.postMessage(cmd)

    const flush = () => {
      if (rafScheduled) return
      rafScheduled = true
      requestAnimationFrame(() => {
        rafScheduled = false
        if (disposed) return
        const arr = [...lines.values()].sort((a, b) => a.rank - b.rank)
        setState((s) => ({ ...s, lines: arr, depth: arr[0]?.depth ?? 0 }))
      })
    }

    const startSearch = (fen: string) => {
      lines.clear()
      side = sideToMove(fen)
      searching = true
      send('position fen ' + fen)
      send('go depth ' + depthRef.current)
      if (!disposed) setState((s) => ({ ...s, thinking: true }))
    }

    analyzeRef.current = (fen: string) => {
      if (searching) {
        pending = fen
        send('stop') // the bestmove handler picks up `pending`
      } else {
        pending = null
        startSearch(fen)
      }
    }
    stopRef.current = () => {
      pending = null
      if (searching) send('stop')
    }

    worker.onerror = () => {
      if (!disposed)
        setState((s) => ({
          ...s,
          unavailable: true,
          ready: false,
          thinking: false,
        }))
    }

    worker.onmessage = (e: MessageEvent) => {
      const text = typeof e.data === 'string' ? e.data : String(e.data ?? '')
      if (!text) return

      if (text === 'readyok') {
        if (!disposed) setState((s) => ({ ...s, ready: true }))
        return
      }
      if (text.startsWith('bestmove')) {
        searching = false
        if (pending) {
          const next = pending
          pending = null
          startSearch(next)
        } else if (!disposed) {
          setState((s) => ({ ...s, thinking: false }))
        }
        return
      }
      if (text.startsWith('info') && text.includes(' pv ')) {
        const parsed = parseInfo(text, side)
        if (parsed) {
          lines.set(parsed.rank, parsed)
          flush()
        }
      }
    }

    send('uci')
    send('setoption name MultiPV value ' + MULTIPV)
    send('isready')

    return () => {
      disposed = true
      analyzeRef.current = () => {}
      stopRef.current = () => {}
      try {
        send('quit')
      } catch {
        /* worker already gone */
      }
      worker.terminate()
    }
  }, [])

  const analyze = useCallback((fen: string) => analyzeRef.current(fen), [])
  const stop = useCallback(() => stopRef.current(), [])

  return { ...state, analyze, stop }
}
