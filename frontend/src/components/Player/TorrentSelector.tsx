'use client'

import { TorrentButton } from './TorrentButton'
import type { Torrent, PlayerState } from '@/hooks/useTorrentStream'

interface TorrentSelectorProps {
  torrents: Torrent[]
  selected: Torrent | null
  state: PlayerState
  onSelect: (torrent: Torrent) => void
}

export function TorrentSelector({ torrents, selected, state, onSelect }: TorrentSelectorProps) {
  return (
    <div className="flex flex-wrap gap-2">
      {torrents.map(torrent => (
        <TorrentButton
          key={torrent.hash}
          torrent={torrent}
          selected={selected}
          state={state}
          onSelect={onSelect}
        />
      ))}
    </div>
  )
}
