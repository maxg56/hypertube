import type { Metadata } from 'next'
import Thumbnails from '@/components/page/Thumbnails'

export const metadata: Metadata = {
  title: 'Accueil — Hypertube',
}

export default function MainPage() {
  return <Thumbnails />
}
