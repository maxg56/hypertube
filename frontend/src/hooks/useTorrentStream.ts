'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { apiClient } from '@/lib/api'

export type PlayerState = 'idle' | 'checking' | 'starting' | 'downloading' | 'streaming' | 'error'

export interface Torrent {
  url: string
  hash: string
  quality: string
  type: string
  size: string
  seeds: number
  peers: number
  magnet: string
}

interface TorrentStatusResponse {
  data?: { status?: string; progress?: number }
}

interface TorrentDownloadResponse {
  data?: { info_hash?: string }
  info_hash?: string
}

export function useTorrentStream(initialTorrent: Torrent | null, movieId: number) {
  const { t } = useTranslation()
  const [selected, setSelected] = useState<Torrent | null>(initialTorrent)
  const [state, setState] = useState<PlayerState>('checking')
  const [progress, setProgress] = useState(0)
  const [infoHash, setInfoHash] = useState<string | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => () => {
    if (pollRef.current) clearInterval(pollRef.current)
  }, [])

  // Fast-path check: is the selected torrent already ready on the server?
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
        const head = await apiClient.head(`/stream/${hash}`)
        if (!cancelled && head.status === 200) {
          setInfoHash(hash)
          setState('streaming')
          return
        }
      } catch { /* fall through to status check */ }
      if (cancelled) return
      try {
        const json = await apiClient.get<TorrentStatusResponse>(`/torrent/status/${hash}`)
        if (cancelled) return
        const { status } = json.data ?? {}
        if (!cancelled && status === 'ready') {
          setInfoHash(hash)
          setState('streaming')
          return
        }
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
        const json = await apiClient.get<TorrentStatusResponse>(`/torrent/status/${hash}`)
        const { status, progress: prog } = json.data ?? {}
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

  const selectTorrent = useCallback((torrent: Torrent) => {
    if (torrent.hash === selected?.hash) return
    stopPolling()
    setInfoHash(null)
    setErrorMsg(null)
    setSelected(torrent)
  }, [selected, stopPolling])

  const startWatch = useCallback(async () => {
    if (!selected) return
    setState('starting')
    setErrorMsg(null)
    try {
      const json = await apiClient.post<TorrentDownloadResponse>('/torrent/download', {
        magnet_uri: selected.magnet,
        movie_id: movieId,
      })
      const hash = (json.data?.info_hash ?? (json as TorrentDownloadResponse).info_hash ?? '').toLowerCase()
      setInfoHash(hash)
      setState('downloading')
      pollStatus(hash)
    } catch {
      setState('error')
      setErrorMsg(t('movie.error_stream'))
    }
  }, [selected, movieId, pollStatus, t])

  return { selected, state, progress, infoHash, errorMsg, selectTorrent, startWatch }
}
