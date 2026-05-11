import { notFound } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { getAccessToken } from '@/lib/session'
import { MoviePlayer } from '@/components/page/MoviePlayer'
import { CommentSection } from '@/components/page/CommentSection'
import { FavoriteButton } from '@/components/page/FavoriteButton'
import { WatchLaterButton } from '@/components/page/WatchLaterButton'
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

export async function generateMetadata({ params }: { params: Promise<{ id: string }> }): Promise<Metadata> {
  const { id } = await params
  const movie = await getMovie(id)
  if (!movie) return { title: 'Film introuvable — Hypertube' }
  return { title: `${movie.title} (${movie.year}) — Hypertube` }
}

function formatRuntime(minutes: number): string {
  if (!minutes) return ''
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return h > 0 ? `${h}h ${m.toString().padStart(2, '0')}m` : `${m}m`
}

export default async function MovieDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const movie = await getMovie(id)
  if (!movie) notFound()

  return (
    <div className="min-h-screen">
      <div className="max-w-5xl mx-auto px-4 pt-4">
        <Link
          href="/"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          <ArrowLeft className="size-4" />
          Retour
        </Link>
      </div>

      {movie.backdrop_url && (
        <div className="relative w-full h-56 sm:h-72 md:h-96 overflow-hidden">
          <img
            src={movie.backdrop_url}
            alt=""
            className="w-full h-full object-cover"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-background via-background/60 to-transparent" />
        </div>
      )}

      <div className="max-w-5xl mx-auto px-4 pb-16">
        <div className={`flex gap-6 ${movie.backdrop_url ? '-mt-24 relative z-10' : 'mt-8'}`}>
          {movie.poster_url && (
            <div className="hidden sm:block shrink-0 w-36 md:w-48">
              <img
                src={movie.poster_url}
                alt={movie.title}
                className="w-full rounded-lg shadow-lg border border-border"
              />
            </div>
          )}

          <div className="flex flex-col gap-3 justify-end min-w-0">
            <h1 className="text-2xl md:text-4xl font-bold leading-tight">{movie.title}</h1>

            <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
              {movie.year && <span>{movie.year}</span>}
              {movie.runtime > 0 && <span>{formatRuntime(movie.runtime)}</span>}
              {movie.rating > 0 && (
                <span className="text-destructive font-semibold">⭐ {movie.rating.toFixed(1)}</span>
              )}
            </div>

            {(movie.genres?.length ?? 0) > 0 && (
              <div className="flex flex-wrap gap-2">
                {movie.genres.map(genre => (
                  <span
                    key={genre}
                    className="text-xs px-2 py-0.5 rounded-full border border-border bg-muted text-muted-foreground"
                  >
                    {genre}
                  </span>
                ))}
              </div>
            )}

            {movie.overview && (
              <p className="text-sm text-muted-foreground leading-relaxed max-w-2xl">{movie.overview}</p>
            )}

            <div className="flex items-center gap-2 flex-wrap">
              <FavoriteButton tmdbId={movie.id} />
              <WatchLaterButton tmdbId={movie.id} />
            </div>
          </div>
        </div>

        {(movie.cast?.length ?? 0) > 0 && (
          <section className="mt-10">
            <h2 className="text-lg font-semibold mb-3">Casting</h2>
            <div className="flex gap-3 overflow-x-auto pb-2">
              {movie.cast.slice(0, 12).map(member => (
                <div
                  key={`${member.name}-${member.order}`}
                  className="shrink-0 w-24 text-center"
                >
                  <div className="w-16 h-16 rounded-full bg-muted mx-auto flex items-center justify-center overflow-hidden">
                    <span className="text-xl font-bold text-muted-foreground">
                      {member.name.charAt(0)}
                    </span>
                  </div>
                  <p className="text-xs font-medium mt-2 truncate">{member.name}</p>
                  <p className="text-xs text-muted-foreground truncate">{member.character}</p>
                </div>
              ))}
            </div>
          </section>
        )}

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
