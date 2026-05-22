import type { MoveClassification } from '../types'

export interface ClassificationMeta {
  label: string
  icon: string
  color: string // text color class
  badge: string // background badge classes
}

/** Display metadata for each move classification, worst → best handled by color. */
export const CLASSIFICATION_META: Record<MoveClassification, ClassificationMeta> = {
  brilliant: {
    label: 'Brilliant',
    icon: '!!',
    color: 'text-cyan-300',
    badge: 'bg-cyan-500/20 text-cyan-300',
  },
  great: {
    label: 'Great',
    icon: '!',
    color: 'text-blue-300',
    badge: 'bg-blue-500/20 text-blue-300',
  },
  best: {
    label: 'Best',
    icon: '★',
    color: 'text-chess-gold',
    badge: 'bg-chess-gold/20 text-chess-gold',
  },
  excellent: {
    label: 'Excellent',
    icon: '✓',
    color: 'text-green-400',
    badge: 'bg-green-500/20 text-green-400',
  },
  good: {
    label: 'Good',
    icon: '·',
    color: 'text-green-500',
    badge: 'bg-green-500/15 text-green-500',
  },
  book: {
    label: 'Book',
    icon: '◆',
    color: 'text-chess-text-muted',
    badge: 'bg-chess-elevated text-chess-text-muted',
  },
  inaccuracy: {
    label: 'Inaccuracy',
    icon: '?!',
    color: 'text-yellow-500',
    badge: 'bg-yellow-500/20 text-yellow-500',
  },
  mistake: {
    label: 'Mistake',
    icon: '?',
    color: 'text-orange-500',
    badge: 'bg-orange-500/20 text-orange-400',
  },
  blunder: {
    label: 'Blunder',
    icon: '??',
    color: 'text-red-500',
    badge: 'bg-red-500/20 text-red-400',
  },
}
