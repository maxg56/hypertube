'use client'
import React from 'react'

export function useFavorites() {
  const [favorites, setFavorites] = React.useState<Set<number>>(new Set())
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    fetch('/api/v1/users/favorites?limit=500', { credentials: 'include' })
      .then((r) => r.json())
      .then((json) => {
        const ids = (json.data?.favorites ?? []).map((f: { tmdb_id: number }) => f.tmdb_id)
        setFavorites(new Set(ids))
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

  return { favorites, loading, isFavorite, toggle }
}
