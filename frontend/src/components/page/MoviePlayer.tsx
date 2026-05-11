'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useProgressSync } from '@/hooks/useProgressSync'
import { useSubtitleTracks } from '@/hooks/useSubtitleTracks'
import { Subtitles } from 'lucide-react'

interface Torrent {
  url: string
  hash: string
  quality: string
  type: string
  size: string
  seeds: number
  peers: number
  magnet: string
}

type PlayerState = 'idle' | 'checking' | 'starting' | 'downloading' | 'streaming' | 'error'

interface MoviePlayerProps {
  torrents: Torrent[]
  movieId: number
}

export function MoviePlayer({ torrents, movieId }: MoviePlayerProps) {
  const { t } = useTranslation()
  const [selected, setSelected] = useState<Torrent | null>(torrents[0] ?? null)
  const [state, setState] = useState<PlayerState>('checking')
  const [progress, setProgress] = useState(0)
  const [infoHash, setInfoHash] = useState<string | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const videoRef = useRef<HTMLVideoElement | null>(null)

  const isStreaming = state === 'streaming'

  useProgressSync(videoRef, movieId, isStreaming)
  const subtitleStatus = useSubtitleTracks(videoRef, movieId, isStreaming)

  useEffect(() => () => {
    if (pollRef.current) clearInterval(pollRef.current)
  }, [])

  // Check if the selected torrent is already ready on the server.
  useEffect(() => {
    if (!selected) {
      setState('idle')
      return
    }
    let cancelled = false
    setState('checking')
    void (async () => {
      const hash = selected.hash.toLowerCase()
      try {
        // Fast path: 200 means the file is on disk and streamable right now.
        // Avoid head.ok (200-299) — 202 means pending, not ready.
        const head = await fetch(`/api/v1/stream/${hash}`, { method: 'HEAD', credentials: 'include' })
        if (!cancelled && head.status === 200) {
          setInfoHash(hash)
          setState('streaming')
          return
        }
      } catch { /* fall through to status check */ }
      if (cancelled) return
      try {
        const res = await fetch(`/api/v1/torrent/status/${hash}`, { credentials: 'include' })
        if (cancelled) return
        if (res.ok) {
          const { status } = ((await res.json()).data ?? {}) as { status?: string }
          if (!cancelled && status === 'ready') {
            setInfoHash(hash)
            setState('streaming')
            return
          }
        }
        // !res.ok (e.g. 404 = never downloaded) → fall through to idle
      } catch { /* fall through */ }
      if (!cancelled) setState('idle')
    })()
    return () => { cancelled = true }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected])

  const stopPolling = useCallback(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current)
      pollRef.current = null
    }
  }, [])

  const pollStatus = useCallback((hash: string) => {
    pollRef.current = setInterval(async () => {
      try {
        const res = await fetch(`/api/v1/torrent/status/${hash}`, { credentials: 'include' })
        if (!res.ok) return
        const { status, progress: prog } = (await res.json()).data ?? {}
        if (status === 'error') {
          stopPolling()
          setState('error')
          setErrorMsg(t('movie.error_stream'))
          return
        }
        setProgress(Math.round((prog ?? 0) * 100))
        if (status === 'pending' || status === 'downloading' || status === 'ready') {
          stopPolling()
          setState('streaming')
        }
      } catch { /* keep polling on transient errors */ }
    }, 2000)
  }, [stopPolling, t])

  const handleTorrentSelect = useCallback((torrent: Torrent) => {
    if (torrent.hash === selected?.hash) return
    stopPolling()
    setInfoHash(null)
    setErrorMsg(null)
    setSelected(torrent)
  }, [selected, stopPolling])

  const handleWatch = useCallback(async () => {
    if (!selected) return
    setState('starting')
    setErrorMsg(null)
    try {
      const res = await fetch('/api/v1/torrent/download', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ magnet_uri: selected.magnet, movie_id: movieId }),
      })
      if (!res.ok) throw new Error(`${res.status}`)
      const json = await res.json()
      const hash = (json.data?.info_hash ?? json.info_hash as string).toLowerCase()
      setInfoHash(hash)
      setState('downloading')
      pollStatus(hash)
    } catch {
      setState('error')
      setErrorMsg(t('movie.error_stream'))
    }
  }, [selected, movieId, pollStatus, t])

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
        <PlayerPlaceholder state={state} progress={progress} errorMsg={errorMsg} t={t} />
      )}

      {isStreaming && subtitleStatus === 'none' && (
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground/60 select-none">
          <span className="relative inline-flex shrink-0">
            <Subtitles className="size-3.5" />
            <span className="absolute inset-0 flex items-center">
              <span className="w-full border-t border-muted-foreground/60 rotate-[-35deg]" />
            </span>
          </span>
          <span>{t('movie.no_subtitles_available')}</span>
        </div>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <div className="flex flex-wrap gap-2">
          {torrents.map(torrent => (
            <TorrentButton
              key={torrent.hash}
              torrent={torrent}
              selected={selected}
              state={state}
              onSelect={handleTorrentSelect}
            />
          ))}
        </div>
        {(state === 'idle' || state === 'error') && (
          <button
            onClick={handleWatch}
            disabled={!selected}
            className="px-4 py-1.5 rounded-md bg-sidebar-primary text-white text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            {t('movie.watch')}
          </button>
        )}
      </div>
    </div>
  )
}

function PlayerPlaceholder({
  state,
  progress,
  errorMsg,
  t,
}: {
  state: PlayerState
  progress: number
  errorMsg: string | null
  t: (key: string, opts?: Record<string, unknown>) => string
}) {
  return (
    <div className="w-full aspect-video rounded-lg bg-muted flex items-center justify-center">
      {(state === 'checking' || state === 'starting') && (
        <span className="text-muted-foreground text-sm animate-pulse">{t('movie.loading')}</span>
      )}
      {state === 'idle' && (
        <span className="text-muted-foreground text-sm">{t('movie.select_quality')}</span>
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
      {state === 'error' && (
        <span className="text-destructive text-sm">{errorMsg}</span>
      )}
    </div>
  )
}

function TorrentButton({
  torrent,
  selected,
  state,
  onSelect,
}: {
  torrent: Torrent
  selected: Torrent | null
  state: PlayerState
  onSelect: (t: Torrent) => void
}) {
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
