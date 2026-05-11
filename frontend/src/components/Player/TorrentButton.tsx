'use client'

import { cn } from '@/lib/utils'
import type { Torrent, PlayerState } from '@/hooks/useTorrentStream'

interface TorrentButtonProps {
  torrent: Torrent
  selected: Torrent | null
  state: PlayerState
  onSelect: (torrent: Torrent) => void
}

export function TorrentButton({ torrent, selected, state, onSelect }: TorrentButtonProps) {
  const isSelected = selected?.hash === torrent.hash
  const locked = state === 'starting'

  return (
    <button
      onClick={() => { if (!locked) onSelect(torrent) }}
      disabled={locked}
      className={cn(
        'text-xs px-3 py-1.5 rounded-md border transition-colors',
        isSelected
          ? 'bg-sidebar-primary text-white border-sidebar-primary'
          : 'bg-muted border-border text-muted-foreground hover:border-sidebar-primary',
        locked && 'opacity-50 cursor-not-allowed',
      )}
    >
      {torrent.quality} {torrent.type && `· ${torrent.type}`}
      {torrent.size && <span className="ml-1 opacity-70">{torrent.size}</span>}
    </button>
  )
}
