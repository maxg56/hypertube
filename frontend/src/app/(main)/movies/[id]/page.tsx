import { notFound } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { getAccessToken } from '@/lib/session'
import { MovieHero } from '@/components/Player/MovieHero'
import { MovieCast } from '@/components/Player/MovieCast'
import { MoviePlayer } from '@/components/Player/MoviePlayer'
import { CommentSection } from '@/components/page/CommentSection'
import type { Metadata } from 'next'

interface CastMember {
  name: string
  character: string
  order: number
}

interface Torrent {
  url: string
  hash: string
  quality: string
  type: string
  size: string
  seeds: number
  peers: number
  magnet: string
}

interface Comment {
  id: number
  user_id: number
  username: string
  avatar_url: string
  content: string
  created_at: string
}

export interface MovieDetail {
  id: number
  imdb_id: string
  title: string
  year: string
  release_date: string
  overview: string
  runtime: number
  rating: number
  poster_url: string
  backdrop_url: string
  genres: string[]
  cast: CastMember[]
  torrents: Torrent[]
  comments: Comment[]
  watched: boolean
}

async function getMovie(id: string): Promise<MovieDetail | null> {
  const token = await getAccessToken()
  try {
    const res = await fetch(`${process.env.API_URL ?? 'http://caddy'}/api/v1/library/movies/${id}`, {
      headers: { Authorization: `Bearer ${token ?? ''}` },
      next: { revalidate: 300 },
    })
    if (res.status === 404) return null
    if (!res.ok) throw new Error(`API error ${res.status}`)
    const json = await res.json()
    return json.data ?? json
  } catch {
    return null
  }
}

function formatRuntime(minutes: number): string {
  if (!minutes) return ''
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return h > 0 ? `${h}h ${m.toString().padStart(2, '0')}m` : `${m}m`
}

export async function generateMetadata({ params }: { params: Promise<{ id: string }> }): Promise<Metadata> {
  const { id } = await params
  const movie = await getMovie(id)
  if (!movie) return { title: 'Film introuvable — Hypertube' }
  return { title: `${movie.title} (${movie.year}) — Hypertube` }
}

export default async function MovieDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const movie = await getMovie(id)
  if (!movie) notFound()

  return (
    <div className="min-h-screen">
      <Link
        href="/"
        className="fixed top-20 left-4 z-50 inline-flex items-center justify-center text-foreground/80 hover:text-foreground bg-background/70 hover:bg-background/90 backdrop-blur-sm border border-border/50 rounded-full p-2 shadow-sm transition-colors"
        aria-label="Retour"
      >
        <ArrowLeft className="size-4" />
      </Link>

      {movie.backdrop_url && (
        <div className="relative w-full h-56 sm:h-72 md:h-96 overflow-hidden">
          <img src={movie.backdrop_url} alt="" className="w-full h-full object-cover" />
          <div className="absolute inset-0 bg-gradient-to-t from-background via-background/60 to-transparent" />
        </div>
      )}

      <div className="max-w-5xl mx-auto px-4 pb-16">
        <MovieHero
          id={movie.id}
          title={movie.title}
          year={movie.year}
          runtime={formatRuntime(movie.runtime)}
          rating={movie.rating}
          overview={movie.overview}
          genres={movie.genres ?? []}
          posterUrl={movie.poster_url}
          backdropUrl={movie.backdrop_url}
        />

        <MovieCast cast={movie.cast ?? []} />

        <section className="mt-10">
          <MoviePlayer torrents={movie.torrents ?? []} movieId={movie.id} />
        </section>

        <section className="mt-12">
          <CommentSection movieId={movie.id} initialComments={movie.comments ?? []} />
        </section>
      </div>
    </div>
  )
}
