'use client'

import { useState, useEffect, useCallback, useRef } from 'react'

export interface Movie {
  id: number
  imdb_id: string
  title: string
  year: string
  rating: number
  poster_url: string
  genres: string[]
}

interface MoviesResponse {
  data: {
    results: Movie[]
    total: number
    next_cursor: string | null
  }
}

export interface MovieFilters {
  query: string
  genre: string
  rating: string
  year: string
  sort_by: string
}

function isValidYear(year: string): boolean {
  if (!/^\d{4}$/.test(year)) return false
  const y = parseInt(year, 10)
  return y >= 1888 && y <= new Date().getFullYear()
}

export function useMovies(filters: MovieFilters) {
  const [movies, setMovies] = useState<Movie[]>([])
  const [nextCursor, setNextCursor] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [initialLoading, setInitialLoading] = useState(true)
  const [hasMore, setHasMore] = useState(true)
  const abortRef = useRef<AbortController | null>(null)

  const fetchMovies = useCallback(async (cursor: string | null, reset: boolean) => {
    if (abortRef.current) abortRef.current.abort()
    const controller = new AbortController()
    abortRef.current = controller

    setLoading(true)

    const params = new URLSearchParams()
    if (filters.query) params.set('q', filters.query)
    if (filters.genre) params.set('genre', filters.genre)
    if (filters.rating) params.set('rating', filters.rating)
    if (isValidYear(filters.year)) params.set('year', filters.year)
    if (filters.sort_by) params.set('sort_by', filters.sort_by)
    if (cursor) params.set('cursor', cursor)

    try {
      const res = await fetch(`/api/v1/library/movies?${params.toString()}`, {
        signal: controller.signal,
        credentials: 'include',
      })
      if (!res.ok) throw new Error('fetch failed')
      const json: MoviesResponse = await res.json()
      const { results, next_cursor } = json.data

      setMovies(prev => {
        if (reset) return results ?? []
        const map = new Map(prev.map(m => [m.id, m]))
        for (const m of results ?? []) map.set(m.id, m)
        return Array.from(map.values())
      })
      setNextCursor(next_cursor ?? null)
      setHasMore(!!next_cursor)
    } catch (err: unknown) {
      if (err instanceof Error && err.name === 'AbortError') return
    } finally {
      setLoading(false)
      setInitialLoading(false)
    }
  }, [filters.query, filters.genre, filters.rating, isValidYear(filters.year) ? filters.year : '', filters.sort_by])

  useEffect(() => {
    setInitialLoading(true)
    setMovies([])
    setNextCursor(null)
    setHasMore(true)
    fetchMovies(null, true)
  }, [fetchMovies])

  const loadMore = useCallback(() => {
    if (!loading && hasMore && nextCursor) {
      fetchMovies(nextCursor, false)
    }
  }, [loading, hasMore, nextCursor, fetchMovies])

  return { movies, loading, initialLoading, hasMore, loadMore }
}
