'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'

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

type PlayerState = 'idle' | 'starting' | 'downloading' | 'streaming' | 'error'

interface MoviePlayerProps {
  torrents: Torrent[]
  movieId: number
}

export function MoviePlayer({ torrents, movieId }: MoviePlayerProps) {
  const { t } = useTranslation()
  const [selected, setSelected] = useState<Torrent | null>(torrents[0] ?? null)
  const [state, setState] = useState<PlayerState>('idle')
  const [progress, setProgress] = useState(0)
  const [infoHash, setInfoHash] = useState<string | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const videoRef = useRef<HTMLVideoElement | null>(null)

  const stopPolling = useCallback(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current)
      pollRef.current = null
    }
  }, [])

  useEffect(() => () => stopPolling(), [stopPolling])

  const pollStatus = useCallback((hash: string) => {
    pollRef.current = setInterval(async () => {
      try {
        const res = await fetch(`/api/v1/torrent/status/${hash}`, { credentials: 'include' })
        if (!res.ok) return
        const json = await res.json()
        const { status, progress: prog } = json.data ?? json

        if (status === 'error') {
          stopPolling()
          setState('error')
          setErrorMsg(t('movie.error_stream'))
          return
        }

        setProgress(Math.round((prog ?? 0) * 100))

        if (status === 'downloading' || status === 'completed') {
          stopPolling()
          setState('streaming')
        }
      } catch {
        // keep polling on transient errors
      }
    }, 2000)
  }, [stopPolling, t])

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
      const hash = json.data?.info_hash ?? json.info_hash
      setInfoHash(hash)
      setState('downloading')
      pollStatus(hash)
    } catch {
      setState('error')
      setErrorMsg(t('movie.error_stream'))
    }
  }, [selected, movieId, pollStatus, t])

  if (!torrents.length) {
    return (
      <p className="text-sm text-muted-foreground">{t('movie.no_torrents')}</p>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {state === 'streaming' && infoHash ? (
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
        <div className="w-full aspect-video rounded-lg bg-muted flex items-center justify-center">
          {state === 'idle' && (
            <span className="text-muted-foreground text-sm">{t('movie.select_quality')}</span>
          )}
          {state === 'starting' && (
            <span className="text-muted-foreground text-sm animate-pulse">{t('movie.loading')}</span>
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
      )}

      <div className="flex flex-wrap items-center gap-3">
        <div className="flex flex-wrap gap-2">
          {torrents.map(torrent => (
            <button
              key={torrent.hash}
              onClick={() => { if (state === 'idle') setSelected(torrent) }}
              disabled={state !== 'idle'}
              className={cn(
                'text-xs px-3 py-1.5 rounded-md border transition-colors',
                selected?.hash === torrent.hash
                  ? 'bg-sidebar-primary text-white border-sidebar-primary'
                  : 'bg-muted border-border text-muted-foreground hover:border-sidebar-primary',
                state !== 'idle' && 'opacity-50 cursor-not-allowed',
              )}
            >
              {torrent.quality} {torrent.type && `· ${torrent.type}`}
              {torrent.size && <span className="ml-1 opacity-70">{torrent.size}</span>}
            </button>
          ))}
        </div>

        {state === 'idle' && (
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
