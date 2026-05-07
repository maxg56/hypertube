'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { MovieCard, MovieCardSkeleton } from '@/components/page/MovieCard'
import { MovieFiltersBar } from '@/components/page/MovieFilters'
import { useMovies, type MovieFilters } from '@/hooks/useMovies'

const SKELETON_COUNT = 20

const DEFAULT_FILTERS: MovieFilters = {
  query: '',
  genre: '',
  rating: '',
  year: '',
  sort_by: 'seeds',
}

export default function Thumbnails() {
  const [filters, setFilters] = useState<MovieFilters>(DEFAULT_FILTERS)
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [watchedMovies] = useState<Set<number>>(new Set())
  const sentinelRef = useRef<HTMLDivElement | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const activeFilters: MovieFilters = { ...filters, query: debouncedQuery }
  const { movies, loading, initialLoading, hasMore, loadMore } = useMovies(activeFilters)

  const handleSearchChange = useCallback((value: string) => {
    setFilters(prev => ({ ...prev, query: value }))
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => setDebouncedQuery(value), 300)
  }, [])

  const handleFilterChange = useCallback((key: keyof Omit<MovieFilters, 'query'>, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }))
  }, [])

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(
      entries => {
        if (entries[0].isIntersecting) loadMore()
      },
      { rootMargin: '200px' },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [loadMore])

  return (
    <div className="flex flex-col gap-4">
      <MovieFiltersBar
        filters={filters}
        onSearchChange={handleSearchChange}
        onFilterChange={handleFilterChange}
      />
      <div className="px-4 pb-6">
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
          {initialLoading
            ? Array.from({ length: SKELETON_COUNT }).map((_, i) => <MovieCardSkeleton key={i} />)
            : movies.map(movie => (
                <MovieCard
                  key={movie.id}
                  movie={movie}
                  watched={watchedMovies.has(movie.id)}
                />
              ))
          }
          {!initialLoading && loading &&
            Array.from({ length: 10 }).map((_, i) => <MovieCardSkeleton key={`more-${i}`} />)
          }
        </div>
        <div ref={sentinelRef} className="h-1" />
        {!loading && !hasMore && movies.length > 0 && (
          <p className="text-center text-sm text-muted-foreground mt-6">—</p>
        )}
      </div>
    </div>
  )
}
