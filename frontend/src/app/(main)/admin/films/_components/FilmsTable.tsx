'use client'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { FilmRow } from './FilmRow'
import type { AdminFilm } from '../hooks/useAdminFilms'

interface FilmsTableProps {
  films: AdminFilm[]
  total: number
  offset: number
  limit: number
  actionId: number | null
  actionType: 'delete' | 'redownload' | null
  onDelete: (film: AdminFilm) => void
  onReDownload: (film: AdminFilm) => void
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
              <th className="text-left py-2 pr-4 font-medium">{t('admin.col_title')}</th>
              <th className="text-left py-2 pr-4 font-medium">{t('admin.col_info_hash')}</th>
              <th className="text-left py-2 pr-4 font-medium">{t('admin.col_language')}</th>
              <th className="text-left py-2 pr-4 font-medium">{t('admin.col_status')}</th>
              <th className="text-right py-2 pr-4 font-medium">{t('admin.col_size')}</th>
              <th className="text-right py-2 pr-4 font-medium">{t('admin.col_watchers')}</th>
              <th className="text-right py-2 font-medium">{t('admin.col_actions')}</th>
            </tr>
          </thead>
          <tbody>
            {films.map((film) => (
              <FilmRow
                key={film.id}
                film={film}
                isActing={actionId === film.id}
                actionType={actionId === film.id ? actionType : null}
                onDelete={() => onDelete(film)}
                onReDownload={() => onReDownload(film)}
              />
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
