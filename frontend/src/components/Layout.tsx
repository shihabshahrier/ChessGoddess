import { ReactNode } from 'react'
import { Header } from './Header'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  return (
    <div className="min-h-screen bg-chess-bg relative">
      <Header />
      <main className="relative z-10">
        {children}
      </main>
    </div>
  )
}
