import { Link } from 'react-router-dom'
import { InteractiveBoard } from '../components/InteractiveBoard'
import { useChessGame } from '../hooks/useChessGame'

// Italian Game — a recognizable, lively position for the hero board.
const HERO_FEN = 'r1bqk1nr/pppp1ppp/2n5/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4'

export function HomePage() {
  const game = useChessGame(HERO_FEN)

  return (
    <div>
      {/* Hero */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-16 pb-20">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <div>
            <span className="inline-block text-xs font-medium tracking-widest uppercase text-chess-gold border border-chess-gold/40 rounded-full px-3 py-1 mb-6">
              Engine + AI chess analysis
            </span>
            <h1 className="font-serif text-5xl md:text-6xl font-bold leading-tight text-chess-text mb-6">
              See chess <span className="text-chess-gold">thinking.</span>
            </h1>
            <p className="text-chess-text-muted text-lg mb-8 max-w-md">
              Stockfish-deep evaluation, move-by-move classification, and
              plain-English explanations — on an interactive board you can
              actually explore.
            </p>
            <div className="flex flex-wrap gap-3">
              <Link
                to="/upload"
                className="bg-chess-gold text-chess-bg px-7 py-3 rounded-lg font-semibold hover:bg-chess-gold-light transition-colors"
              >
                Start an Analysis
              </Link>
              <Link
                to="/analysis"
                className="border border-chess-border text-chess-text px-7 py-3 rounded-lg font-semibold hover:border-chess-gold hover:text-chess-gold transition-colors"
              >
                Open the Board
              </Link>
            </div>
            <p className="mt-6 text-sm text-chess-text-dim">
              Paste a PGN · scan a board photo · drop a FEN
            </p>
          </div>

          <div className="w-full max-w-md mx-auto lg:mx-0 lg:ml-auto">
            <InteractiveBoard
              fen={game.fen}
              onMove={game.makeMove}
              getLegalMoves={game.getLegalMoves}
              arrows={[{ from: 'f3', to: 'g5', kind: 'primary' }]}
              disabled
            />
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-24">
        <h2 className="font-serif text-2xl font-semibold text-chess-text mb-8">
          What you get
        </h2>
        <div className="grid md:grid-cols-3 gap-6">
          <FeatureCard
            icon="🔍"
            title="Deep Engine Analysis"
            description="Stockfish with multi-line search. Every move scored by centipawn loss and classified — brilliant to blunder — with per-side accuracy."
          />
          <FeatureCard
            icon="📷"
            title="Scan Any Board"
            description="Photograph an over-the-board game or screenshot a position. Vision models read it into a FEN you can analyze in one click — PNG, JPEG, or WebP."
          />
          <FeatureCard
            icon="🤖"
            title="AI Explanations"
            description="Not just a number. Plain-English reasoning for why a move works — or why it throws the game — powered by language models."
          />
        </div>
      </section>

      {/* Closing CTA */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-24">
        <div className="bg-chess-surface border border-chess-border rounded-2xl p-10 text-center">
          <h2 className="font-serif text-2xl font-semibold text-chess-text mb-3">
            Bring a game in.
          </h2>
          <p className="text-chess-text-muted mb-6 max-w-lg mx-auto">
            A PGN, a board photo, or a single position — start exploring in seconds.
          </p>
          <Link
            to="/upload"
            className="inline-block bg-chess-gold text-chess-bg px-8 py-3 rounded-lg font-semibold hover:bg-chess-gold-light transition-colors"
          >
            New Analysis
          </Link>
        </div>
      </section>
    </div>
  )
}

function FeatureCard({
  icon,
  title,
  description,
}: {
  icon: string
  title: string
  description: string
}) {
  return (
    <div className="bg-chess-surface border border-chess-border rounded-xl p-6 hover:border-chess-gold/50 transition-colors">
      <div className="text-3xl mb-4">{icon}</div>
      <h3 className="font-serif text-lg font-semibold text-chess-text mb-2">
        {title}
      </h3>
      <p className="text-chess-text-muted text-sm leading-relaxed">
        {description}
      </p>
    </div>
  )
}
