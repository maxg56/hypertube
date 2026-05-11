'use client'
import React from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Heart } from 'lucide-react'
import { MovieCard } from '@/components/page/MovieCard'
import type { Movie } from '@/hooks/useMovies'

interface FavoriteMovie {
  tmdb_id: number
  title: string
  poster_url: string
  rating: number
  language: string
  release_date: string
}

function toMovie(f: FavoriteMovie): Movie {
  return {
    id: f.tmdb_id,
    imdb_id: '',
    title: f.title,
    year: f.release_date?.slice(0, 4) ?? '',
    rating: f.rating,
    poster_url: f.poster_url,
    genres: [],
  }
}

export default function FavoritesPage() {
  const { t } = useTranslation()
  const [movies, setMovies] = React.useState<FavoriteMovie[]>([])
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    fetch('/api/v1/users/favorites?limit=100', { credentials: 'include' })
      .then((r) => r.json())
      .then((json) => setMovies(json.data?.favorites ?? []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold mb-6 flex items-center gap-2">
        <Heart className="size-6 text-destructive fill-current" />
        {t('favorites.title')}
      </h1>

      {loading ? (
        <div className="flex justify-center py-24">
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </div>
      ) : movies.length === 0 ? (
        <p className="text-muted-foreground text-sm py-16 text-center">{t('favorites.empty')}</p>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
          {movies.map((f) => (
            <MovieCard key={f.tmdb_id} movie={toMovie(f)} watched={false} />
          ))}
        </div>
      )}
    </div>
  )
}
