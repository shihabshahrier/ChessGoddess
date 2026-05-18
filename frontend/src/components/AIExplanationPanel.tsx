import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'

interface AIExplanationPanelProps {
  moveId: string
  classification?: string
  evaluation?: number
  onExplain: (moveId: string) => Promise<string>
}

export function AIExplanationPanel({ moveId, classification, evaluation, onExplain }: AIExplanationPanelProps) {
  const [explanation, setExplanation] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [isOpen, setIsOpen] = useState(false)

  const handleExplain = async () => {
    setLoading(true)
    try {
      const result = await onExplain(moveId)
      setExplanation(result)
      setIsOpen(true)
    } catch (err) {
      console.error('Failed to get explanation:', err)
    } finally {
      setLoading(false)
    }
  }

  const isBlunder = classification === 'blunder'
  const isMistake = classification === 'mistake'

  return (
    <div className="mt-4">
      <button
        onClick={handleExplain}
        disabled={loading}
        className={`w-full py-3 px-4 rounded-lg font-medium transition-all ${
          isBlunder 
            ? 'bg-red-500/20 text-red-400 hover:bg-red-500/30' 
            : isMistake
            ? 'bg-orange-500/20 text-orange-400 hover:bg-orange-500/30'
            : 'bg-chess-gold/20 text-chess-gold hover:bg-chess-gold/30'
        } disabled:opacity-50`}
      >
        {loading ? (
          <span className="flex items-center justify-center gap-2">
            <span className="animate-spin">◌</span>
            AI is thinking...
          </span>
        ) : isBlunder ? (
          '💡 Why is this a blunder?'
        ) : isMistake ? (
          '💡 Why is this a mistake?'
        ) : (
          '💡 Explain this move'
        )}
      </button>

      <AnimatePresence>
        {isOpen && explanation && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="mt-4 bg-chess-elevated border border-chess-border rounded-lg p-4"
          >
            <div className="flex items-center gap-2 mb-2">
              <span className="text-chess-gold">🤖</span>
              <span className="text-sm font-medium text-chess-gold">AI Coach</span>
            </div>
            <p className="text-chess-text-muted text-sm leading-relaxed">
              {explanation}
            </p>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
