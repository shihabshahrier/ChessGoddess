/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        chess: {
          bg: '#1a1a1a',
          surface: '#242424',
          elevated: '#2d2d2d',
          border: '#3a3a3a',
          gold: '#d4af37',
          'gold-light': '#e8c84a',
          'gold-dark': '#b8960f',
          walnut: '#5c4033',
          'walnut-light': '#7a5c47',
          charcoal: '#2c2c2c',
          text: '#e8e8e8',
          'text-muted': '#9a9a9a',
          'text-dim': '#6a6a6a',
        }
      },
      fontFamily: {
        serif: ['Playfair Display', 'Georgia', 'serif'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'eval-swing': 'evalSwing 0.6s cubic-bezier(0.34, 1.56, 0.64, 1)',
        'board-shake': 'boardShake 0.3s ease-in-out',
        'piece-glide': 'pieceGlide 0.4s ease-out',
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.4s ease-out',
      },
      keyframes: {
        evalSwing: {
          '0%': { transform: 'translateY(-100%)' },
          '100%': { transform: 'translateY(0)' },
        },
        boardShake: {
          '0%, 100%': { transform: 'translateX(0)' },
          '25%': { transform: 'translateX(-2px)' },
          '75%': { transform: 'translateX(2px)' },
        },
        pieceGlide: {
          '0%': { transform: 'translate(var(--from-x), var(--from-y))', opacity: '0.8' },
          '100%': { transform: 'translate(0, 0)', opacity: '1' },
        },
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}
