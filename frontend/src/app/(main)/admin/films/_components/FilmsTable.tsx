'use client'
import React from 'react'
import Link from 'next/link'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { FilmRow } from './FilmRow'
import type { AdminGroupedFilm } from '../hooks/useAdminFilms'

interface FilmsTableProps {
  films: AdminGroupedFilm[]
  total: number
  offset: number
  limit: number
  actionId: number | null
  actionType: 'delete' | 'redownload' | null
  onDelete: (torrentId: number) => void
  onReDownload: (torrentId: number) => void
  onPageChange: (offset: number) => void
}

export function FilmsTable({
  films, total, offset, limit, actionId, actionType,
  onDelete, onReDownload, onPageChange,
}: FilmsTableProps) {
  const { t } = useTranslation()

  return (
    <>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b text-muted-foreground">
              <th className="text-left py-2 pl-6 pr-4 font-medium">{t('admin.col_info_hash')}</th>
              <th className="text-left py-2 pr-4 font-medium">{t('admin.col_language')}</th>
              <th className="text-left py-2 pr-4 font-medium">{t('admin.col_status')}</th>
              <th className="text-right py-2 pr-4 font-medium">{t('admin.col_size')}</th>
              <th className="text-right py-2 font-medium">{t('admin.col_actions')}</th>
            </tr>
          </thead>
          <tbody>
            {films.map((film) => (
              <React.Fragment key={film.movie_id}>
                <tr className="bg-muted/20 border-b">
                  <td colSpan={5} className="py-2.5 px-3">
                    <div className="flex items-center justify-between gap-4">
                      {film.tmdb_id ? (
                        <Link
                          href={`/movies/${film.tmdb_id}`}
                          className="font-semibold hover:underline hover:text-sidebar-primary truncate"
                        >
                          {film.title || `movie_${film.movie_id}`}
                        </Link>
                      ) : (
                        <span className="font-semibold truncate">{film.title || `movie_${film.movie_id}`}</span>
                      )}
                      <span className="text-xs text-muted-foreground shrink-0">
                        {film.watchers_count} {t('admin.watchers')}
                      </span>
                    </div>
                  </td>
                </tr>
                {film.torrents.map((torrent) => (
                  <FilmRow
                    key={torrent.id}
                    torrent={torrent}
                    isActing={actionId === torrent.id}
                    actionType={actionId === torrent.id ? actionType : null}
                    onDelete={() => onDelete(torrent.id)}
                    onReDownload={() => onReDownload(torrent.id)}
                  />
                ))}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>

      {total > limit && (
        <div className="flex items-center justify-between mt-4 text-sm text-muted-foreground">
          <span>{offset + 1}–{Math.min(offset + limit, total)} / {total}</span>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={offset === 0}
              onClick={() => onPageChange(Math.max(0, offset - limit))}
            >
              {t('admin.pagination_prev')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={offset + limit >= total}
              onClick={() => onPageChange(offset + limit)}
            >
              {t('admin.pagination_next')}
            </Button>
          </div>
        </div>
      )}
    </>
  )
}
