'use client'
import { Bookmark } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useWatchLater } from '@/hooks/useWatchLater'

interface WatchLaterButtonProps {
  tmdbId: number
  className?: string
}

export function WatchLaterButton({ tmdbId, className }: WatchLaterButtonProps) {
  const { t } = useTranslation()
  const { isInWatchLater, toggle, loading } = useWatchLater()

  const active = isInWatchLater(tmdbId)

  return (
    <button
      onClick={() => toggle(tmdbId)}
      disabled={loading}
      title={active ? t('watch_later.remove') : t('watch_later.add')}
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-sm font-medium transition-colors',
        'border border-border hover:bg-muted disabled:opacity-50',
        active ? 'text-sidebar-primary' : 'text-muted-foreground hover:text-sidebar-primary',
        className,
      )}
    >
      <Bookmark className={cn('size-4', active && 'fill-current')} />
      {active ? t('watch_later.remove') : t('watch_later.add')}
    </button>
  )
}
