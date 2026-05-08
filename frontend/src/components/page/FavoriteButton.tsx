'use client'
import { Heart } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useFavorites } from '@/hooks/useFavorites'

interface FavoriteButtonProps {
  tmdbId: number
  className?: string
}

export function FavoriteButton({ tmdbId, className }: FavoriteButtonProps) {
  const { t } = useTranslation()
  const { isFavorite, toggle, loading } = useFavorites()

  const active = isFavorite(tmdbId)

  return (
    <button
      onClick={() => toggle(tmdbId)}
      disabled={loading}
      title={active ? t('favorites.remove') : t('favorites.add')}
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-sm font-medium transition-colors',
        'border border-border hover:bg-muted disabled:opacity-50',
        active ? 'text-destructive' : 'text-muted-foreground hover:text-destructive',
        className,
      )}
    >
      <Heart className={cn('size-4', active && 'fill-current')} />
      {active ? t('favorites.remove') : t('favorites.add')}
    </button>
  )
}
