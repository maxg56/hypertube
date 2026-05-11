'use client'

import { useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useTorrentStream } from '@/hooks/useTorrentStream'
import { useProgressSync } from '@/hooks/useProgressSync'
import { useSubtitleTracks } from '@/hooks/useSubtitleTracks'
import { PlayerPlaceholder } from './PlayerPlaceholder'
import { TorrentSelector } from './TorrentSelector'
import { SubtitleIndicator } from './SubtitleIndicator'
import type { Torrent } from '@/hooks/useTorrentStream'

export type { Torrent }

interface MoviePlayerProps {
  torrents: Torrent[]
  movieId: number
}

export function MoviePlayer({ torrents, movieId }: MoviePlayerProps) {
  const { t } = useTranslation()
  const videoRef = useRef<HTMLVideoElement | null>(null)

  const { selected, state, progress, infoHash, errorMsg, selectTorrent, startWatch } =
    useTorrentStream(torrents[0] ?? null, movieId)

  const isStreaming = state === 'streaming'
  useProgressSync(videoRef, movieId, isStreaming)
  const subtitleStatus = useSubtitleTracks(videoRef, movieId, isStreaming)

  if (!torrents.length) {
    return <p className="text-sm text-muted-foreground">{t('movie.no_torrents')}</p>
  }

  return (
    <div className="flex flex-col gap-4">
      {isStreaming && infoHash ? (
        <video
          ref={videoRef}
          src={`/api/v1/stream/${infoHash}`}
          controls
          preload="metadata"
          crossOrigin="use-credentials"
          className="w-full rounded-lg bg-black aspect-video"
        >
          {t('movie.video_unsupported')}
        </video>
      ) : (
        <PlayerPlaceholder
          state={state}
          progress={progress}
          errorMsg={errorMsg}
          hasSelected={!!selected}
          onWatch={startWatch}
        />
      )}

      <SubtitleIndicator show={isStreaming && subtitleStatus === 'none'} />

      <TorrentSelector
        torrents={torrents}
        selected={selected}
        state={state}
        onSelect={selectTorrent}
      />
    </div>
  )
}
