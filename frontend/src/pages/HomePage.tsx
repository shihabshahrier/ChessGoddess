import { Link } from 'react-router-dom'

export function HomePage() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
      <div className="text-center">
        <h1 className="font-serif text-5xl md:text-6xl font-bold text-chess-text mb-6">
          See Chess <span className="text-chess-gold">Thinking</span>
        </h1>
        <p className="text-chess-text-muted text-xl max-w-2xl mx-auto mb-12">
          A cinematic chess analysis studio that turns raw engine output into 
          readable insight, visual tension, and beautiful review experiences.
        </p>
        
        <div className="flex justify-center gap-4">
          <Link 
            to="/upload" 
            className="bg-chess-gold text-chess-bg px-8 py-3 rounded-lg font-semibold hover:bg-chess-gold-light transition-colors"
          >
            Analyze a Game
          </Link>
          <button className="border border-chess-border text-chess-text px-8 py-3 rounded-lg font-semibold hover:border-chess-gold hover:text-chess-gold transition-colors">
            Watch Demo
          </button>
        </div>
      </div>
      
      <div className="mt-24 grid md:grid-cols-3 gap-8">
        <FeatureCard 
          icon="🔍"
          title="Deep Analysis"
          description="Stockfish-powered evaluation with move classification and insights"
        />
        <FeatureCard 
          icon="🎬"
          title="Cinematic Review"
          description="Beautiful animated reviews with eval bar physics and piece glide"
        />
        <FeatureCard 
          icon="🤖"
          title="AI Explanations"
          description="Human-readable explanations powered by advanced language models"
        />
      </div>
    </div>
  )
}

function FeatureCard({ icon, title, description }: { icon: string; title: string; description: string }) {
  return (
    <div className="bg-chess-surface border border-chess-border rounded-xl p-6 hover:border-chess-gold/50 transition-colors">
      <div className="text-4xl mb-4">{icon}</div>
      <h3 className="font-serif text-xl font-semibold text-chess-text mb-2">{title}</h3>
      <p className="text-chess-text-muted">{description}</p>
    </div>
  )
}
