import { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Header } from './Header'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  return (
    <div className="min-h-screen flex flex-col bg-chess-bg relative">
      <Header />
      <main className="relative z-10 flex-1">{children}</main>
      <footer className="relative z-10 border-t border-chess-border mt-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex flex-wrap items-center justify-between gap-3 text-sm text-chess-text-dim">
          <span>
            <span className="text-chess-gold">♔</span> ChessGoddess — engine + AI
            analysis
          </span>
          <nav className="flex gap-5">
            <Link to="/analysis" className="hover:text-chess-gold transition-colors">
              Analysis Board
            </Link>
            <Link to="/upload" className="hover:text-chess-gold transition-colors">
              New Analysis
            </Link>
          </nav>
        </div>
      </footer>
    </div>
  )
}
