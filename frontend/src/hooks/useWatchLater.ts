'use client'
import React from 'react'
import { apiClient } from '@/lib/api'

export interface WatchLaterMovie {
  tmdb_id: number
  title: string
  poster_url: string
  rating: number
  language: string
  release_date: string
}

export function useWatchLater() {
  const [items, setItems] = React.useState<Set<number>>(new Set())
  const [list, setList] = React.useState<WatchLaterMovie[]>([])
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    apiClient.get<{ data: { items: WatchLaterMovie[] } }>('/users/watch-later?limit=500')
      .then((json) => {
        const movies = json.data?.items ?? []
        setList(movies)
        setItems(new Set(movies.map((m) => m.tmdb_id)))
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const isInWatchLater = React.useCallback((tmdbId: number) => items.has(tmdbId), [items])

  const toggle = React.useCallback(async (tmdbId: number) => {
    const was = items.has(tmdbId)

    setItems((prev) => {
      const next = new Set(prev)
      was ? next.delete(tmdbId) : next.add(tmdbId)
      return next
    })
    setList((prev) =>
      was ? prev.filter((m) => m.tmdb_id !== tmdbId) : prev
    )

    try {
      if (was) {
        await apiClient.delete(`/users/watch-later/${tmdbId}`)
      } else {
        await apiClient.post('/users/watch-later', { tmdb_id: tmdbId })
      }
    } catch {
      setItems((prev) => {
        const next = new Set(prev)
        was ? next.add(tmdbId) : next.delete(tmdbId)
        return next
      })
    }
  }, [items])

  return { items, list, loading, isInWatchLater, toggle }
}
