'use client'

import { useState, useRef, useEffect } from 'react'
import { Settings2, Check } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import type { Torrent, PlayerState } from '@/hooks/useTorrentStream'

interface TorrentSelectorProps {
  torrents: Torrent[]
  selected: Torrent | null
  state: PlayerState
  onSelect: (torrent: Torrent) => void
}

export function TorrentSelector({ torrents, selected, state, onSelect }: TorrentSelectorProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    function onClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [open])

  const locked = state === 'starting'
  const label = selected
    ? `${selected.quality}${selected.type ? ` · ${selected.type}` : ''}`
    : t('movie.select_quality')

  return (
    <div ref={ref} className="relative inline-flex">
      <button
        onClick={() => setOpen(v => !v)}
        disabled={locked}
        className={cn(
          'flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-white/20 bg-black/50 text-white backdrop-blur-sm hover:border-white/50 transition-colors',
          open && 'border-white/50',
          locked && 'opacity-50 cursor-not-allowed',
        )}
      >
        <Settings2 className="size-3.5" />
        <span>{label}</span>
      </button>

      {open && (
        <div className="absolute bottom-full mb-2 left-0 w-56 rounded-lg border border-border bg-background shadow-lg overflow-hidden z-10">
          <p className="px-3 py-2 text-xs font-semibold text-muted-foreground border-b border-border">
            {t('movie.quality')}
          </p>
          <ul>
            {torrents.map(torrent => {
              const isSelected = selected?.hash === torrent.hash
              return (
                <li key={torrent.hash}>
                  <button
                    onClick={() => { onSelect(torrent); setOpen(false) }}
                    className={cn(
                      'w-full flex items-center justify-between px-3 py-2.5 text-sm hover:bg-muted transition-colors',
                      isSelected ? 'text-sidebar-primary font-medium' : 'text-foreground',
                    )}
                  >
                    <span>
                      {torrent.quality}
                      {torrent.type && (
                        <span className="ml-1.5 text-xs text-muted-foreground">· {torrent.type}</span>
                      )}
                    </span>
                    <span className="flex items-center gap-2 shrink-0">
                      {torrent.size && (
                        <span className="text-xs text-muted-foreground">{torrent.size}</span>
                      )}
                      {isSelected && <Check className="size-3.5 text-sidebar-primary" />}
                    </span>
                  </button>
                </li>
              )
            })}
          </ul>
        </div>
      )}
    </div>
  )
}
