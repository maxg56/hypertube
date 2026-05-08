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

type PlayerState = 'idle' | 'checking' | 'starting' | 'downloading' | 'streaming' | 'error'

interface MoviePlayerProps {
  torrents: Torrent[]
  movieId: number
}

const SAVE_INTERVAL_MS = 5000

export function MoviePlayer({ torrents, movieId }: MoviePlayerProps) {
  const { t, i18n } = useTranslation()
  const [selected, setSelected] = useState<Torrent | null>(torrents[0] ?? null)
  const [state, setState] = useState<PlayerState>('checking')
  const [progress, setProgress] = useState(0)
  const [infoHash, setInfoHash] = useState<string | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const saveRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const videoRef = useRef<HTMLVideoElement | null>(null)

  const stopPolling = useCallback(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current)
      pollRef.current = null
    }
  }, [])

  const stopSaving = useCallback(() => {
    if (saveRef.current) {
      clearInterval(saveRef.current)
      saveRef.current = null
    }
  }, [])

  useEffect(() => () => { stopPolling(); stopSaving() }, [stopPolling, stopSaving])

  // Restore saved position once the video element is ready.
  const restorePosition = useCallback(async () => {
    const video = videoRef.current
    if (!video) return
    try {
      const res = await fetch(`/api/v1/movies/${movieId}/progress`, { credentials: 'include' })
      if (!res.ok) return
      const json = await res.json()
      const sec: number = (json.data ?? json).progress_sec ?? 0
      if (sec > 0) video.currentTime = sec
    } catch {
      // ignore — just start from the beginning
    }
  }, [movieId])

  // Periodically save the current playback position.
  const startSaving = useCallback(() => {
    saveRef.current = setInterval(async () => {
      const video = videoRef.current
      if (!video || video.paused || video.ended) return
      const sec = Math.floor(video.currentTime)
      if (sec <= 0) return
      try {
        await fetch(`/api/v1/movies/${movieId}/progress`, {
          method: 'PUT',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ progress_sec: sec }),
        })
      } catch {
        // ignore transient save errors
      }
    }, SAVE_INTERVAL_MS)
  }, [movieId])

  // Check if the selected torrent is already ready on the server.
  useEffect(() => {
    if (!selected) {
      setState('idle')
      return
    }
    let cancelled = false
    setState('checking')
    void (async () => {
      try {
        const hash = selected.hash.toLowerCase()
        const res = await fetch(`/api/v1/torrent/status/${hash}`, { credentials: 'include' })
        if (cancelled) return
        if (res.ok) {
          const json = await res.json()
          const { status } = (json.data ?? json) as { status: string }
          if (!cancelled && status === 'ready') {
            setInfoHash(hash)
            setState('streaming')
            return
          }
        }
      } catch {
        // fall through to idle
      }
      if (!cancelled) setState('idle')
    })()
    return () => { cancelled = true }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected])

  // When streaming starts, restore position and begin saving.
  useEffect(() => {
    if (state !== 'streaming') return
    stopSaving()
    // Delay slightly so the video element has time to mount.
    const timer = setTimeout(() => {
      void restorePosition()
      startSaving()
    }, 300)
    return () => clearTimeout(timer)
  }, [state, restorePosition, startSaving, stopSaving])

  // Fetch subtitle files with credentials and inject them as blob-URL tracks.
  // This bypasses browser inconsistencies with the declarative <track> + default
  // attribute, and ensures the JWT cookie is sent with each subtitle request.
  useEffect(() => {
    const video = videoRef.current
    if (!video || state !== 'streaming') return

    const userLang = i18n.language.slice(0, 2)
    const langs = userLang === 'en' ? ['en'] : [userLang, 'en']
    const blobUrls: string[] = []
    let cancelled = false

    langs.forEach(async (lang) => {
      try {
        const res = await fetch(`/api/v1/movies/${movieId}/subtitles/${lang}`, {
          credentials: 'include',
        })
        if (!res.ok || cancelled) return
        const blob = await res.blob()
        if (cancelled) return

        const url = URL.createObjectURL(blob)
        blobUrls.push(url)

        const el = document.createElement('track')
        el.kind = 'subtitles'
        el.src = url
        el.srclang = lang
        el.label = t(`movie.subtitle_${lang}`)
        video.appendChild(el)

        if (lang === userLang) {
          const tracks = video.textTracks
          for (let i = 0; i < tracks.length; i++) {
            if (tracks[i].language === lang) {
              tracks[i].mode = 'showing'
              break
            }
          }
        }
      } catch {
        // subtitle not available for this language — silently ignore
      }
    })

    return () => {
      cancelled = true
      Array.from(video.querySelectorAll('track')).forEach(el => el.remove())
      blobUrls.forEach(url => URL.revokeObjectURL(url))
    }
  }, [state, movieId, i18n.language, t])

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

        if (status === 'downloading' || status === 'ready') {
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
      )}

      <div className="flex flex-wrap items-center gap-3">
        <div className="flex flex-wrap gap-2">
          {torrents.map(torrent => (
            <button
              key={torrent.hash}
              onClick={() => { if (state === 'idle' || state === 'checking') setSelected(torrent) }}
              disabled={state !== 'idle' && state !== 'checking'}
              className={cn(
                'text-xs px-3 py-1.5 rounded-md border transition-colors',
                selected?.hash === torrent.hash
                  ? 'bg-sidebar-primary text-white border-sidebar-primary'
                  : 'bg-muted border-border text-muted-foreground hover:border-sidebar-primary',
                state !== 'idle' && state !== 'checking' && 'opacity-50 cursor-not-allowed',
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
