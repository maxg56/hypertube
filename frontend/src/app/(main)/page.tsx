import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Accueil — Hypertube',
}

export default function HomePage() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-2">Films</h1>
      <p className="text-muted-foreground">Le catalogue arrive bientôt…</p>
    </div>
  )
}
