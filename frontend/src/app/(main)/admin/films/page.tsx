'use client'
import React from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Loader2, Trash2, RefreshCw } from 'lucide-react'

interface AdminFilm {
  id: number
  movie_id: number
  tmdb_id: number
  info_hash: string
  status: string
  file_path: string
  file_size: number
  downloaded: number
  progress: number
  title: string
  poster_path: string
  created_at: string
  watchers_count: number
  watcher_ids: number[]
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '—'
  const units = ['B', 'KB', 'MB', 'GB']
  let v = bytes
  let i = 0
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++ }
  return `${v.toFixed(1)} ${units[i]}`
}

interface ConfirmDialogProps {
  title: string
  body: string
  cancelLabel: string
  confirmLabel: string
  onCancel: () => void
  onConfirm: () => void
}

function ConfirmDialog({ title, body, cancelLabel, confirmLabel, onCancel, onConfirm }: ConfirmDialogProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
      <div className="bg-card border border-border rounded-xl shadow-lg p-6 max-w-sm w-full mx-4">
        <h2 className="text-base font-semibold mb-2">{title}</h2>
        <p className="text-sm text-muted-foreground mb-6">{body}</p>
        <div className="flex gap-3 justify-end">
          <Button variant="outline" size="sm" onClick={onCancel}>{cancelLabel}</Button>
          <Button variant="destructive" size="sm" onClick={onConfirm}>{confirmLabel}</Button>
        </div>
      </div>
    </div>
  )
}

export default function AdminFilmsPage() {
  const { t } = useTranslation()
  const [films, setFilms] = React.useState<AdminFilm[]>([])
  const [total, setTotal] = React.useState(0)
  const [offset, setOffset] = React.useState(0)
  const [loading, setLoading] = React.useState(true)
  const [pendingDelete, setPendingDelete] = React.useState<AdminFilm | null>(null)
  const [actionId, setActionId] = React.useState<number | null>(null)
  const [actionType, setActionType] = React.useState<'delete' | 'redownload' | null>(null)
  const limit = 20

  const fetchFilms = React.useCallback(async (off: number) => {
    setLoading(true)
    try {
      const res = await fetch(`/api/v1/admin/films?limit=${limit}&offset=${off}`, { credentials: 'include' })
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

  const handleDeleteConfirm = async () => {
    if (!pendingDelete) return
    const id = pendingDelete.id
    setPendingDelete(null)
    setActionId(id)
    setActionType('delete')
    try {
      await fetch(`/api/v1/admin/films/${id}`, { method: 'DELETE', credentials: 'include' })
      setFilms((prev) => prev.filter((f) => f.id !== id))
      setTotal((prev) => prev - 1)
    } finally {
      setActionId(null)
      setActionType(null)
    }
  }

  const handleReDownload = async (film: AdminFilm) => {
    setActionId(film.id)
    setActionType('redownload')
    try {
      await fetch(`/api/v1/admin/films/${film.id}/download`, { method: 'POST', credentials: 'include' })
      await fetchFilms(offset)
    } finally {
      setActionId(null)
      setActionType(null)
    }
  }

  const statusLabel = (s: string) => {
    const map: Record<string, string> = {
      ready: t('admin.status_ready'),
      downloading: t('admin.status_downloading'),
      pending: t('admin.status_pending'),
      error: t('admin.status_error'),
    }
    return map[s] ?? s
  }

  const statusClass = (s: string) => {
    const map: Record<string, string> = {
      ready: 'bg-green-500/15 text-green-600 dark:text-green-400',
      downloading: 'bg-sidebar-primary/15 text-sidebar-primary',
      pending: 'bg-muted text-muted-foreground',
      error: 'bg-destructive/15 text-destructive',
    }
    return map[s] ?? 'bg-muted text-muted-foreground'
  }

  return (
    <>
      {pendingDelete && (
        <ConfirmDialog
          title={t('admin.confirm_delete_title')}
          body={t('admin.confirm_delete_body')}
          cancelLabel={t('admin.confirm_delete_cancel')}
          confirmLabel={t('admin.confirm_delete_confirm')}
          onCancel={() => setPendingDelete(null)}
          onConfirm={handleDeleteConfirm}
        />
      )}

      <Card>
        <CardHeader>
          <CardTitle>{t('admin.films_title')}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="size-6 animate-spin text-muted-foreground" />
            </div>
          ) : films.length === 0 ? (
            <p className="text-muted-foreground text-sm py-8 text-center">{t('admin.no_films')}</p>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b text-muted-foreground">
                      <th className="text-left py-2 pr-4 font-medium">{t('admin.col_title')}</th>
                      <th className="text-left py-2 pr-4 font-medium">{t('admin.col_status')}</th>
                      <th className="text-right py-2 pr-4 font-medium">{t('admin.col_size')}</th>
                      <th className="text-right py-2 pr-4 font-medium">{t('admin.col_watchers')}</th>
                      <th className="text-right py-2 font-medium">{t('admin.col_actions')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {films.map((film) => {
                      const isActing = actionId === film.id
                      return (
                        <tr key={film.id} className="border-b last:border-0 hover:bg-muted/40 transition-colors">
                          <td className="py-2.5 pr-4">
                            <div className="flex items-center gap-3">
                              {film.poster_path && (
                                <img
                                  src={`https://image.tmdb.org/t/p/w92${film.poster_path}`}
                                  alt={film.title}
                                  className="h-10 w-7 rounded object-cover shrink-0"
                                />
                              )}
                              <div>
                                <p className="font-medium leading-tight">{film.title || film.info_hash.slice(0, 12)}</p>
                                {film.status === 'downloading' && (
                                  <div className="mt-1 flex items-center gap-2">
                                    <div className="h-1 w-24 rounded-full bg-muted overflow-hidden">
                                      <div
                                        className="h-full bg-sidebar-primary transition-all"
                                        style={{ width: `${Math.min(100, film.progress)}%` }}
                                      />
                                    </div>
                                    <span className="text-xs text-muted-foreground">{film.progress.toFixed(0)}%</span>
                                  </div>
                                )}
                              </div>
                            </div>
                          </td>
                          <td className="py-2.5 pr-4">
                            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${statusClass(film.status)}`}>
                              {statusLabel(film.status)}
                            </span>
                          </td>
                          <td className="py-2.5 pr-4 text-right tabular-nums text-muted-foreground">
                            {formatBytes(film.file_size)}
                          </td>
                          <td className="py-2.5 pr-4 text-right tabular-nums">{film.watchers_count}</td>
                          <td className="py-2.5 text-right">
                            <div className="flex items-center justify-end gap-2">
                              <Button
                                variant="outline"
                                size="sm"
                                disabled={isActing}
                                onClick={() => handleReDownload(film)}
                                title={t('admin.action_redownload')}
                              >
                                {isActing && actionType === 'redownload'
                                  ? <Loader2 className="size-3.5 animate-spin" />
                                  : <RefreshCw className="size-3.5" />
                                }
                              </Button>
                              <Button
                                variant="destructive"
                                size="sm"
                                disabled={isActing}
                                onClick={() => setPendingDelete(film)}
                                title={t('admin.action_delete')}
                              >
                                {isActing && actionType === 'delete'
                                  ? <Loader2 className="size-3.5 animate-spin" />
                                  : <Trash2 className="size-3.5" />
                                }
                              </Button>
                            </div>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>

              {total > limit && (
                <div className="flex items-center justify-between mt-4 text-sm text-muted-foreground">
                  <span>{offset + 1}–{Math.min(offset + limit, total)} / {total}</span>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" disabled={offset === 0} onClick={() => goTo(Math.max(0, offset - limit))}>
                      {t('admin.pagination_prev')}
                    </Button>
                    <Button variant="outline" size="sm" disabled={offset + limit >= total} onClick={() => goTo(offset + limit)}>
                      {t('admin.pagination_next')}
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </>
  )
}
