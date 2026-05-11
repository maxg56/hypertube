'use client'
import React from 'react'

export interface FavoriteMovie {
  tmdb_id: number
  title: string
  poster_url: string
  rating: number
  language: string
  release_date: string
}

export function useFavorites() {
  const [favorites, setFavorites] = React.useState<Set<number>>(new Set())
  const [list, setList] = React.useState<FavoriteMovie[]>([])
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    fetch('/api/v1/users/favorites?limit=500', { credentials: 'include' })
      .then((r) => r.json())
      .then((json) => {
        const movies: FavoriteMovie[] = json.data?.favorites ?? []
        setList(movies)
        setFavorites(new Set(movies.map((f) => f.tmdb_id)))
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const isFavorite = React.useCallback((tmdbId: number) => favorites.has(tmdbId), [favorites])

  const toggle = React.useCallback(async (tmdbId: number) => {
    const was = favorites.has(tmdbId)

    setFavorites((prev) => {
      const next = new Set(prev)
      was ? next.delete(tmdbId) : next.add(tmdbId)
      return next
    })
    setList((prev) => was ? prev.filter((m) => m.tmdb_id !== tmdbId) : prev)

    try {
      if (was) {
        await fetch(`/api/v1/users/favorites/${tmdbId}`, { method: 'DELETE', credentials: 'include' })
      } else {
        await fetch('/api/v1/users/favorites', {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ tmdb_id: tmdbId }),
        })
      }
    } catch {
      setFavorites((prev) => {
        const next = new Set(prev)
        was ? next.add(tmdbId) : next.delete(tmdbId)
        return next
      })
    }
  }, [favorites])

  return { favorites, list, loading, isFavorite, toggle }
}
