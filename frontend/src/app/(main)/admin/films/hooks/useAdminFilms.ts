'use client'
import React from 'react'

export interface AdminTorrent {
  id: number
  info_hash: string
  status: string
  file_size: number
  downloaded: number
  progress: number
  quality: string
  created_at: string
}

export interface AdminGroupedFilm {
  movie_id: number
  tmdb_id: number
  title: string
  poster_path: string
  watchers_count: number
  watcher_ids: number[]
  torrents: AdminTorrent[]
}

const LIMIT = 20

export function useAdminFilms() {
  const [films, setFilms] = React.useState<AdminGroupedFilm[]>([])
  const [total, setTotal] = React.useState(0)
  const [offset, setOffset] = React.useState(0)
  const [loading, setLoading] = React.useState(true)
  const [actionId, setActionId] = React.useState<number | null>(null)
  const [actionType, setActionType] = React.useState<'delete' | 'redownload' | null>(null)

  const fetchFilms = React.useCallback(async (off: number) => {
    setLoading(true)
    try {
      const res = await fetch(`/api/v1/admin/films?limit=${LIMIT}&offset=${off}`, { credentials: 'include' })
      if (!res.ok) return
      const json = await res.json()
      setFilms(json.data?.films ?? [])
      setTotal(json.data?.pagination?.total ?? 0)
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchFilms(0) }, [fetchFilms])

  const goTo = (off: number) => {
    setOffset(off)
    fetchFilms(off)
  }

  const deleteFilm = async (torrentId: number) => {
    setActionId(torrentId)
    setActionType('delete')
    try {
      await fetch(`/api/v1/admin/films/${torrentId}`, { method: 'DELETE', credentials: 'include' })
      await fetchFilms(offset)
    } finally {
      setActionId(null)
      setActionType(null)
    }
  }

  const reDownload = async (torrentId: number) => {
    setActionId(torrentId)
    setActionType('redownload')
    try {
      await fetch(`/api/v1/admin/films/${torrentId}/download`, { method: 'POST', credentials: 'include' })
      await fetchFilms(offset)
    } finally {
      setActionId(null)
      setActionType(null)
    }
  }

  return { films, total, offset, loading, actionId, actionType, limit: LIMIT, goTo, deleteFilm, reDownload }
}
