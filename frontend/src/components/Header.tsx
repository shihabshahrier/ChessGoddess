import { Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'

export function Header() {
  const { user, loading, signIn, signOut } = useAuth()

  return (
    <header className="relative z-20 border-b border-chess-border bg-chess-surface/80 backdrop-blur-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <Link to="/" className="flex items-center gap-2">
            <span className="text-2xl">♔</span>
            <span className="font-serif text-xl font-semibold text-chess-gold">ChessGoddess</span>
          </Link>

          <nav className="flex items-center gap-6">
            <Link to="/analysis" className="text-chess-text-muted hover:text-chess-gold transition-colors">
              Analysis Board
            </Link>
            <Link to="/upload" className="text-chess-text-muted hover:text-chess-gold transition-colors">
              New Analysis
            </Link>

            {loading ? (
              <div className="w-20 h-9" />
            ) : user ? (
              <div className="flex items-center gap-3">
                {user.avatar_url && (
                  <img
                    src={user.avatar_url}
                    alt={user.name}
                    referrerPolicy="no-referrer"
                    className="w-8 h-8 rounded-full border border-chess-border"
                  />
                )}
                <span className="hidden sm:block text-chess-text text-sm">
                  {user.name}
                </span>
                <button
                  onClick={signOut}
                  className="text-chess-text-muted hover:text-chess-gold text-sm transition-colors"
                >
                  Sign Out
                </button>
              </div>
            ) : (
              <button
                onClick={signIn}
                className="bg-chess-gold text-chess-bg px-4 py-2 rounded-md font-medium hover:bg-chess-gold-light transition-colors"
              >
                Sign In
              </button>
            )}
          </nav>
        </div>
      </div>
    </header>
  )
}
