'use client'
import React from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Loader2 } from 'lucide-react'
import { useAdminFilms } from './hooks/useAdminFilms'
import { FilmsTable } from './_components/FilmsTable'

function ConfirmDialog({
  title, body, cancelLabel, confirmLabel, onCancel, onConfirm,
}: {
  title: string
  body: string
  cancelLabel: string
  confirmLabel: string
  onCancel: () => void
  onConfirm: () => void
}) {
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
  const { films, total, offset, loading, actionId, actionType, limit, goTo, deleteFilm, reDownload } = useAdminFilms()
  const [pendingDelete, setPendingDelete] = React.useState<number | null>(null)

  const handleDeleteConfirm = async () => {
    if (pendingDelete === null) return
    const id = pendingDelete
    setPendingDelete(null)
    await deleteFilm(id)
  }

  return (
    <>
      {pendingDelete !== null && (
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
            <FilmsTable
              films={films}
              total={total}
              offset={offset}
              limit={limit}
              actionId={actionId}
              actionType={actionType}
              onDelete={setPendingDelete}
              onReDownload={reDownload}
              onPageChange={goTo}
            />
          )}
        </CardContent>
      </Card>
    </>
  )
}
