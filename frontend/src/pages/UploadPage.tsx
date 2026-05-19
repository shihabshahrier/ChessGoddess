import { useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { uploadGame, createAnalysis } from '../api/games'

export function UploadPage() {
  const navigate = useNavigate()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [dragActive, setDragActive] = useState(false)
  const [pgnInput, setPgnInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true)
    } else if (e.type === 'dragleave') {
      setDragActive(false)
    }
  }

  const readPGNFile = (file: File) => {
    if (!file.name.endsWith('.pgn') && file.type !== 'application/x-chess-pgn') {
      setError('Please upload a .pgn file')
      return
    }
    if (file.size > 1024 * 1024) {
      setError('File too large (max 1MB)')
      return
    }
    const reader = new FileReader()
    reader.onload = (e) => {
      const text = e.target?.result
      if (typeof text === 'string') {
        setPgnInput(text)
        setError('')
      }
    }
    reader.onerror = () => setError('Failed to read file')
    reader.readAsText(file)
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)
    const file = e.dataTransfer.files?.[0]
    if (file) readPGNFile(file)
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) readPGNFile(file)
  }

  const handleAnalyze = async () => {
    if (!pgnInput.trim()) return
    setLoading(true)
    setError('')
    try {
      const { game_id } = await uploadGame(pgnInput)
      const { session_id } = await createAnalysis(game_id)
      navigate(`/analysis/${session_id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed')
      setLoading(false)
    }
  }

  return (
    <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
      <h1 className="font-serif text-3xl font-bold text-chess-text mb-8">Upload a Game</h1>

      {error && (
        <div className="mb-6 p-3 bg-red-900/30 border border-red-700 rounded-lg text-red-300 text-sm">
          {error}
        </div>
      )}

      <div
        className={`border-2 border-dashed rounded-xl p-12 text-center transition-colors ${
          dragActive ? 'border-chess-gold bg-chess-gold/5' : 'border-chess-border'
        }`}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
      >
        <div className="text-5xl mb-4">📁</div>
        <p className="text-chess-text-muted mb-4">
          Drag and drop a PGN file here, or paste below
        </p>
        <input
          ref={fileInputRef}
          type="file"
          accept=".pgn"
          className="hidden"
          id="pgn-upload"
          onChange={handleFileChange}
        />
        <label
          htmlFor="pgn-upload"
          className="inline-block bg-chess-elevated text-chess-text px-6 py-2 rounded-lg cursor-pointer hover:bg-chess-border transition-colors"
        >
          Browse Files
        </label>
      </div>

      <div className="mt-8">
        <label className="block text-chess-text font-medium mb-2">
          Or paste PGN directly
        </label>
        <textarea
          value={pgnInput}
          onChange={(e) => setPgnInput(e.target.value)}
          placeholder="1. e4 e5 2. Nf3 Nc6 ..."
          className="w-full h-48 bg-chess-surface border border-chess-border rounded-lg p-4 text-chess-text font-mono resize-none focus:border-chess-gold focus:outline-none"
        />
      </div>

      <button
        onClick={handleAnalyze}
        disabled={!pgnInput.trim() || loading}
        className="mt-6 w-full bg-chess-gold text-chess-bg py-3 rounded-lg font-semibold hover:bg-chess-gold-light transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {loading ? 'Analyzing...' : 'Analyze Game'}
      </button>
    </div>
  )
}
