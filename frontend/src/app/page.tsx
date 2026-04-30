import type { Metadata } from 'next'
import HomeClient from './(main)/home-client'

export const metadata: Metadata = {
  title: 'Accueil — Hypertube',
}

export default function HomePage() {
  return <HomeClient />
}
