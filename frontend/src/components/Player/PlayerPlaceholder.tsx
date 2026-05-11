'use client'

import { Download } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { PlayerState } from '@/hooks/useTorrentStream'

interface PlayerPlaceholderProps {
  state: PlayerState
  progress: number
  errorMsg: string | null
  hasSelected: boolean
  onWatch: () => void
}

export function PlayerPlaceholder({ state, progress, errorMsg, hasSelected, onWatch }: PlayerPlaceholderProps) {
  const { t } = useTranslation()

  return (
    <div className="w-full h-full bg-muted flex items-center justify-center">
      {(state === 'checking' || state === 'starting') && (
        <span className="text-muted-foreground text-sm animate-pulse">{t('movie.loading')}</span>
      )}
      {(state === 'idle' || state === 'error') && (
        <div className="flex flex-col items-center gap-4">
          {state === 'error' && (
            <span className="text-destructive text-sm">{errorMsg}</span>
          )}
          {state === 'idle' && (
            <Download className="size-10 text-muted-foreground/40" />
          )}
          <button
            onClick={onWatch}
            disabled={!hasSelected}
            className="px-6 py-2 rounded-md bg-sidebar-primary text-white text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            {t('movie.watch')}
          </button>
        </div>
      )}
      {state === 'downloading' && (
        <div className="flex flex-col items-center gap-3 w-48">
          <span className="text-sm text-muted-foreground">
            {t('movie.progress', { percent: progress })}
          </span>
          <div className="w-full h-2 bg-border rounded-full overflow-hidden">
            <div
              className="h-full bg-sidebar-primary transition-all duration-500"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}
    </div>
  )
}
