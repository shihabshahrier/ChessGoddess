import { Link } from 'react-router-dom'

export function Header() {
  return (
    <header className="relative z-20 border-b border-chess-border bg-chess-surface/80 backdrop-blur-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <Link to="/" className="flex items-center gap-2">
            <span className="text-2xl">♔</span>
            <span className="font-serif text-xl font-semibold text-chess-gold">ChessLens</span>
          </Link>
          
          <nav className="flex items-center gap-6">
            <Link to="/upload" className="text-chess-text-muted hover:text-chess-gold transition-colors">
              Upload Game
            </Link>
            <button className="bg-chess-gold text-chess-bg px-4 py-2 rounded-md font-medium hover:bg-chess-gold-light transition-colors">
              Sign In
            </button>
          </nav>
        </div>
      </div>
    </header>
  )
}
