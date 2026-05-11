'use client'

import { useTranslation } from 'react-i18next'
import { Search, Bookmark, Heart, Users } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { MovieFilters } from '@/hooks/useMovies'

const GENRES = [
  'Action', 'Adventure', 'Animation', 'Biography', 'Comedy', 'Crime',
  'Documentary', 'Drama', 'Family', 'Fantasy', 'History', 'Horror',
  'Music', 'Mystery', 'Romance', 'Sci-Fi', 'Thriller', 'War', 'Western',
]

const SORT_OPTIONS = [
  { value: 'seeds', labelKey: 'library.sort_seeds' },
  { value: 'rating', labelKey: 'library.sort_rating' },
  { value: 'year', labelKey: 'library.sort_year' },
  { value: 'title', labelKey: 'library.sort_title' },
  { value: 'download_count', labelKey: 'library.sort_downloads' },
]

interface MovieFiltersProps {
  filters: MovieFilters
  onSearchChange: (value: string) => void
  onFilterChange: (key: keyof Omit<MovieFilters, 'query'>, value: string) => void
  watchLater: boolean
  onWatchLaterChange: (v: boolean) => void
  favorites: boolean
  onFavoritesChange: (v: boolean) => void
}

export function MovieFiltersBar({
  filters, onSearchChange, onFilterChange, watchLater, onWatchLaterChange, favorites, onFavoritesChange,
}: MovieFiltersProps) {
  const { t } = useTranslation()
  const isUserSearch = filters.query.startsWith('@')

  return (
    <div className="flex flex-col gap-3 p-4 pb-0" suppressHydrationWarning>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground pointer-events-none" />
        <input
          type="text"
          value={filters.query}
          onChange={e => onSearchChange(e.target.value)}
          placeholder={t('library.search_placeholder')}
          disabled={watchLater || favorites}
          className="w-full pl-9 pr-4 py-2 bg-card border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-sidebar-primary/50 disabled:opacity-40"
        />
        {isUserSearch && (
          <span className="absolute right-3 top-1/2 -translate-y-1/2 inline-flex items-center gap-1 text-xs font-medium text-sidebar-primary bg-sidebar-primary/10 rounded-full px-2 py-0.5">
            <Users className="size-3" />
            {t('user_search.mode_hint')}
          </span>
        )}
      </div>
      <div className="flex flex-wrap gap-2 items-center">
        <div className={cn('contents', (watchLater || favorites || isUserSearch) && 'opacity-40 pointer-events-none')}>
          <select
            value={filters.genre}
            onChange={e => onFilterChange('genre', e.target.value)}
            className="bg-card border rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-sidebar-primary/50"
          >
            <option value="">{t('library.filter_all_genres')}</option>
            {GENRES.map(g => (
              <option key={g} value={g}>{g}</option>
            ))}
          </select>

          <select
            value={filters.rating}
            onChange={e => onFilterChange('rating', e.target.value)}
            className="bg-card border rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-sidebar-primary/50"
          >
            <option value="">{t('library.filter_all_ratings')}</option>
            {[9, 8, 7, 6, 5].map(r => (
              <option key={r} value={String(r)}>{t('library.filter_rating_min', { rating: r })}</option>
            ))}
          </select>

          <input
            type="number"
            value={filters.year}
            onChange={e => onFilterChange('year', e.target.value)}
            placeholder={t('library.filter_year_placeholder')}
            min={1900}
            max={new Date().getFullYear()}
            className="bg-card border rounded-lg px-3 py-1.5 text-sm w-24 focus:outline-none focus:ring-2 focus:ring-sidebar-primary/50"
          />

          <select
            value={filters.sort_by}
            onChange={e => onFilterChange('sort_by', e.target.value)}
            className="bg-card border rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-sidebar-primary/50"
          >
            {SORT_OPTIONS.map(o => (
              <option key={o.value} value={o.value}>{t(o.labelKey)}</option>
            ))}
          </select>
        </div>

        <div className={cn('contents', isUserSearch && 'opacity-40 pointer-events-none')}>
        <button
          onClick={() => { onFavoritesChange(!favorites); if (!favorites) onWatchLaterChange(false) }}
          className={cn(
            'inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium border transition-colors',
            favorites
              ? 'bg-destructive/10 border-destructive text-destructive'
              : 'bg-card border-border text-muted-foreground hover:text-destructive hover:border-destructive',
          )}
        >
          <Heart className={cn('size-4', favorites && 'fill-current')} />
          {t('favorites.filter')}
        </button>
        <button
          onClick={() => { onWatchLaterChange(!watchLater); if (!watchLater) onFavoritesChange(false) }}
          className={cn(
            'inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium border transition-colors',
            watchLater
              ? 'bg-sidebar-primary/10 border-sidebar-primary text-sidebar-primary'
              : 'bg-card border-border text-muted-foreground hover:text-sidebar-primary hover:border-sidebar-primary',
          )}
        >
          <Bookmark className={cn('size-4', watchLater && 'fill-current')} />
          {t('watch_later.filter')}
        </button>
        </div>
      </div>
    </div>
  )
}
