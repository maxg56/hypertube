'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { MovieCard, MovieCardSkeleton } from '@/components/page/MovieCard'
import { MovieFiltersBar } from '@/components/page/MovieFilters'
import { useMovies, type MovieFilters } from '@/hooks/useMovies'
import { useWatchLater } from '@/hooks/useWatchLater'
import type { Movie } from '@/hooks/useMovies'
import type { WatchLaterMovie } from '@/hooks/useWatchLater'

const SKELETON_COUNT = 20

const DEFAULT_FILTERS: MovieFilters = {
  query: '',
  genre: '',
  rating: '',
  year: '',
  sort_by: 'seeds',
}

function watchLaterToMovie(m: WatchLaterMovie): Movie {
  return {
    id: m.tmdb_id,
    imdb_id: '',
    title: m.title,
    year: m.release_date?.slice(0, 4) ?? '',
    rating: m.rating,
    poster_url: m.poster_url,
    genres: [],
  }
}

export default function Thumbnails() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<MovieFilters>(DEFAULT_FILTERS)
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [watchedMovies] = useState<Set<number>>(new Set())
  const [watchLater, setWatchLater] = useState(false)
  const sentinelRef = useRef<HTMLDivElement | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const activeFilters: MovieFilters = { ...filters, query: debouncedQuery }
  const { movies, loading, initialLoading, hasMore, loadMore } = useMovies(activeFilters)
  const { list: watchLaterList, loading: watchLaterLoading } = useWatchLater()

  const handleSearchChange = useCallback((value: string) => {
    setFilters(prev => ({ ...prev, query: value }))
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => setDebouncedQuery(value), 300)
  }, [])

  const handleFilterChange = useCallback((key: keyof Omit<MovieFilters, 'query'>, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }))
  }, [])

  useEffect(() => {
    if (watchLater) return
    const sentinel = sentinelRef.current
    if (!sentinel) return
    const observer = new IntersectionObserver(
      entries => { if (entries[0].isIntersecting) loadMore() },
      { rootMargin: '200px' },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [loadMore, watchLater])

  const watchLaterMovies = watchLaterList.map(watchLaterToMovie)

  return (
    <div className="flex flex-col gap-4">
      <MovieFiltersBar
        filters={filters}
        onSearchChange={handleSearchChange}
        onFilterChange={handleFilterChange}
        watchLater={watchLater}
        onWatchLaterChange={setWatchLater}
      />
      <div className="px-4 pb-6">
        {watchLater ? (
          watchLaterLoading ? (
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
              {Array.from({ length: 10 }).map((_, i) => <MovieCardSkeleton key={i} />)}
            </div>
          ) : watchLaterMovies.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground py-16">{t('watch_later.empty')}</p>
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
              {watchLaterMovies.map(movie => (
                <MovieCard key={movie.id} movie={movie} watched={watchedMovies.has(movie.id)} />
              ))}
            </div>
          )
        ) : (
          <>
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
              {initialLoading
                ? Array.from({ length: SKELETON_COUNT }).map((_, i) => <MovieCardSkeleton key={i} />)
                : movies.map(movie => (
                    <MovieCard key={movie.id} movie={movie} watched={watchedMovies.has(movie.id)} />
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
          </>
        )}
      </div>
    </div>
  )
}
