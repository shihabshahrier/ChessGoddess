import { useState } from 'react'

export function UploadPage() {
  const [dragActive, setDragActive] = useState(false)
  const [pgnInput, setPgnInput] = useState('')

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true)
    } else if (e.type === 'dragleave') {
      setDragActive(false)
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)
    // TODO: Handle file upload
  }

  const handleAnalyze = () => {
    // TODO: Submit PGN for analysis
    console.log('Analyzing:', pgnInput)
  }

  return (
    <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
      <h1 className="font-serif text-3xl font-bold text-chess-text mb-8">Upload a Game</h1>
      
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
        <input type="file" accept=".pgn" className="hidden" id="pgn-upload" />
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
        disabled={!pgnInput}
        className="mt-6 w-full bg-chess-gold text-chess-bg py-3 rounded-lg font-semibold hover:bg-chess-gold-light transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Analyze Game
      </button>
    </div>
  )
}
