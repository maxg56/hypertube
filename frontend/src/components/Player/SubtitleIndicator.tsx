'use client'

import { Subtitles } from 'lucide-react'
import { useTranslation } from 'react-i18next'

interface SubtitleIndicatorProps {
  show: boolean
}

export function SubtitleIndicator({ show }: SubtitleIndicatorProps) {
  const { t } = useTranslation()

  if (!show) return null

  return (
    <div className="flex items-center gap-1.5 text-xs text-muted-foreground/60 select-none">
      <span className="relative inline-flex shrink-0">
        <Subtitles className="size-3.5" />
        <span className="absolute inset-0 flex items-center">
          <span className="w-full border-t border-muted-foreground/60 rotate-[-35deg]" />
        </span>
      </span>
      <span>{t('movie.no_subtitles_available')}</span>
    </div>
  )
}
