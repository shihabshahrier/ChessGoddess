import { useState, useRef, useCallback, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { InteractiveBoard } from '../components/InteractiveBoard'
import { useChessGame, isValidFen } from '../hooks/useChessGame'
import { useAuth } from '../hooks/useAuth'
import { uploadGame, createAnalysis } from '../api/games'
import { imageToFEN } from '../api/vision'

type Tab = 'pgn' | 'image' | 'fen'
const MAX_IMAGE_BYTES = 10 * 1024 * 1024

const TABS: { id: Tab; label: string; hint: string }[] = [
  { id: 'pgn', label: 'Paste PGN', hint: 'Analyze a full game' },
  { id: 'image', label: 'Scan a Board', hint: 'Photo or screenshot' },
  { id: 'fen', label: 'Enter FEN', hint: 'Jump to a position' },
]

export function UploadPage() {
  const navigate = useNavigate()
  const { user, signIn } = useAuth()
  const game = useChessGame()
  const { reset, makeMove, getLegalMoves } = game

  const [tab, setTab] = useState<Tab>('pgn')
  const [error, setError] = useState('')

  // ── PGN tab ────────────────────────────────────────────────
  const [pgn, setPgn] = useState('')
  const [analyzing, setAnalyzing] = useState(false)
  const [dragActive, setDragActive] = useState(false)

  const readPgnFile = useCallback((file: File) => {
    if (!file.name.endsWith('.pgn') && file.type !== 'application/x-chess-pgn') {
      setError('That is not a .pgn file')
      return
    }
    if (file.size > 1024 * 1024) {
      setError('PGN too large (max 1MB)')
      return
    }
    const reader = new FileReader()
    reader.onload = (e) => {
      if (typeof e.target?.result === 'string') {
        setPgn(e.target.result)
        setError('')
      }
    }
    reader.onerror = () => setError('Could not read the file')
    reader.readAsText(file)
  }, [])

  const handleAnalyzePgn = useCallback(async () => {
    if (!pgn.trim()) return
    setAnalyzing(true)
    setError('')
    try {
      const { game_id } = await uploadGame(pgn)
      const { session_id } = await createAnalysis(game_id)
      navigate(`/review/${session_id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Upload failed')
      setAnalyzing(false)
    }
  }, [pgn, navigate])

  // ── Image tab ──────────────────────────────────────────────
  const [loadingImage, setLoadingImage] = useState(false)
  const [previewURL, setPreviewURL] = useState<string | null>(null)
  const [fen, setFen] = useState('')
  const [hasPosition, setHasPosition] = useState(false)

  const processImage = useCallback(
    async (file: File) => {
      if (!file.type.startsWith('image/')) {
        setError('Choose a PNG, JPEG, or WebP image')
        return
      }
      if (file.size > MAX_IMAGE_BYTES) {
        setError('Image too large (max 10MB)')
        return
      }
      setError('')
      setLoadingImage(true)
      setPreviewURL((prev) => {
        if (prev) URL.revokeObjectURL(prev)
        return URL.createObjectURL(file)
      })
      try {
        const res = await imageToFEN(file)
        setFen(res.fen)
        reset(res.fen)
        setHasPosition(true)
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Could not read the board')
        setHasPosition(false)
      } finally {
        setLoadingImage(false)
      }
    },
    [reset],
  )

  // Paste a screenshot from the clipboard while on the image tab.
  useEffect(() => {
    if (tab !== 'image') return
    const onPaste = (e: ClipboardEvent) => {
      const item = Array.from(e.clipboardData?.items ?? []).find((i) =>
        i.type.startsWith('image/'),
      )
      const file = item?.getAsFile()
      if (file) processImage(file)
    }
    window.addEventListener('paste', onPaste)
    return () => window.removeEventListener('paste', onPaste)
  }, [tab, processImage])

  useEffect(() => {
    return () => {
      if (previewURL) URL.revokeObjectURL(previewURL)
    }
  }, [previewURL])

  // ── FEN tab ────────────────────────────────────────────────
  const handleFenChange = (value: string) => {
    setFen(value)
    if (isValidFen(value)) {
      reset(value)
      setHasPosition(true)
    } else {
      setHasPosition(false)
    }
  }

  const fenOk = isValidFen(fen)

  // ── Shared drag handlers ───────────────────────────────────
  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'dragenter' || e.type === 'dragover') setDragActive(true)
    else if (e.type === 'dragleave') setDragActive(false)
  }
  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)
    const file = e.dataTransfer.files?.[0]
    if (!file) return
    if (tab === 'pgn') readPgnFile(file)
    else if (tab === 'image') processImage(file)
  }

  const switchTab = (next: Tab) => {
    setTab(next)
    setError('')
  }

  const pgnFileRef = useRef<HTMLInputElement>(null)
  const imageFileRef = useRef<HTMLInputElement>(null)

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <h1 className="font-serif text-3xl font-bold text-chess-text mb-1">
        New Analysis
      </h1>
      <p className="text-chess-text-muted mb-8">
        Bring a game or position in — paste a PGN, scan a board photo, or drop a FEN.
      </p>

      {/* Tabs */}
      <div className="flex gap-1 mb-8 border-b border-chess-border">
        {TABS.map((t) => (
          <button
            key={t.id}
            onClick={() => switchTab(t.id)}
            className={`px-5 py-3 text-sm font-medium border-b-2 -mb-px transition-colors ${
              tab === t.id
                ? 'border-chess-gold text-chess-gold'
                : 'border-transparent text-chess-text-muted hover:text-chess-text'
            }`}
          >
            {t.label}
            <span className="block text-xs text-chess-text-dim font-normal mt-0.5">
              {t.hint}
            </span>
          </button>
        ))}
      </div>

      {error && (
        <div className="mb-6 p-3 bg-red-900/30 border border-red-700 rounded-lg text-red-300 text-sm">
          {error}
        </div>
      )}

      {/* ── PGN ── */}
      {tab === 'pgn' && (
        <div className="max-w-3xl">
          <div
            className={`border-2 border-dashed rounded-xl p-8 text-center transition-colors ${
              dragActive ? 'border-chess-gold bg-chess-gold/5' : 'border-chess-border'
            }`}
            onDragEnter={handleDrag}
            onDragLeave={handleDrag}
            onDragOver={handleDrag}
            onDrop={handleDrop}
          >
            <div className="text-4xl mb-3">♟</div>
            <p className="text-chess-text-muted mb-4">
              Drop a <span className="text-chess-text">.pgn</span> file here
            </p>
            <input
              ref={pgnFileRef}
              type="file"
              accept=".pgn"
              className="hidden"
              onChange={(e) => {
                const f = e.target.files?.[0]
                if (f) readPgnFile(f)
              }}
            />
            <button
              onClick={() => pgnFileRef.current?.click()}
              className="bg-chess-elevated text-chess-text px-5 py-2 rounded-lg hover:bg-chess-border transition-colors"
            >
              Browse Files
            </button>
          </div>

          <label className="block text-chess-text font-medium mt-6 mb-2">
            Or paste PGN
          </label>
          <textarea
            value={pgn}
            onChange={(e) => setPgn(e.target.value)}
            placeholder="1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 ..."
            className="w-full h-48 bg-chess-surface border border-chess-border rounded-lg p-4 text-chess-text font-mono text-sm resize-none focus:border-chess-gold focus:outline-none"
          />
          {user ? (
            <button
              onClick={handleAnalyzePgn}
              disabled={!pgn.trim() || analyzing}
              className="mt-5 w-full bg-chess-gold text-chess-bg py-3 rounded-lg font-semibold hover:bg-chess-gold-light transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {analyzing ? 'Uploading…' : 'Analyze Game'}
            </button>
          ) : (
            <div className="mt-5">
              <button
                onClick={signIn}
                className="w-full bg-chess-gold text-chess-bg py-3 rounded-lg font-semibold hover:bg-chess-gold-light transition-colors"
              >
                Sign in to Analyze a Game
              </button>
              <p className="text-chess-text-dim text-xs mt-2 text-center">
                Full-game review is saved to your account. Scanning a board or
                entering a FEN needs no sign-in.
              </p>
            </div>
          )}
        </div>
      )}

      {/* ── Image / FEN — input + live board preview ── */}
      {(tab === 'image' || tab === 'fen') && (
        <div className="grid lg:grid-cols-2 gap-8">
          <div>
            {tab === 'image' && (
              <div
                className={`border-2 border-dashed rounded-xl p-8 text-center transition-colors ${
                  dragActive
                    ? 'border-chess-gold bg-chess-gold/5'
                    : 'border-chess-border'
                }`}
                onDragEnter={handleDrag}
                onDragLeave={handleDrag}
                onDragOver={handleDrag}
                onDrop={handleDrop}
              >
                {previewURL ? (
                  <img
                    src={previewURL}
                    alt="Uploaded board"
                    className="max-h-48 mx-auto rounded-lg mb-4 object-contain"
                  />
                ) : (
                  <div className="text-4xl mb-3">📷</div>
                )}
                <p className="text-chess-text-muted mb-1">
                  {loadingImage
                    ? 'Reading the position…'
                    : 'Drag & drop, paste (⌘V), or browse'}
                </p>
                <p className="text-chess-text-dim text-xs mb-4">
                  PNG, JPEG, or WebP
                </p>
                <input
                  ref={imageFileRef}
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  className="hidden"
                  onChange={(e) => {
                    const f = e.target.files?.[0]
                    if (f) processImage(f)
                  }}
                />
                <button
                  onClick={() => imageFileRef.current?.click()}
                  className="bg-chess-elevated text-chess-text px-5 py-2 rounded-lg hover:bg-chess-border transition-colors"
                >
                  Browse Files
                </button>
              </div>
            )}

            <label className="block text-chess-text font-medium mt-6 mb-2 text-sm">
              {tab === 'image' ? 'Recognized FEN — edit if a piece is off' : 'FEN'}
            </label>
            <input
              value={fen}
              onChange={(e) => handleFenChange(e.target.value)}
              placeholder="rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
              className={`w-full bg-chess-surface border rounded-lg p-3 text-chess-text font-mono text-sm focus:outline-none ${
                fen === '' || fenOk
                  ? 'border-chess-border focus:border-chess-gold'
                  : 'border-red-600'
              }`}
            />
            {fen !== '' && !fenOk && (
              <p className="text-red-400 text-xs mt-1">Not a legal position.</p>
            )}
          </div>

          <div>
            <p className="text-chess-text-muted text-sm mb-3">
              {hasPosition ? 'Preview — drag pieces to correct it' : 'Preview'}
            </p>
            <InteractiveBoard
              fen={game.fen}
              onMove={makeMove}
              getLegalMoves={getLegalMoves}
              disabled={!hasPosition}
            />
            <button
              onClick={() => navigate('/analysis', { state: { fen: game.fen } })}
              disabled={!hasPosition || !fenOk}
              className="mt-6 w-full bg-chess-gold text-chess-bg py-3 rounded-lg font-semibold hover:bg-chess-gold-light transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >
              Open in Analysis Board →
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
