'use client'
import { useTranslation } from 'react-i18next'

const STATUS_CLASSES: Record<string, string> = {
  ready:       'bg-green-500/15 text-green-600 dark:text-green-400',
  downloading: 'bg-sidebar-primary/15 text-sidebar-primary',
  pending:     'bg-muted text-muted-foreground',
  error:       'bg-destructive/15 text-destructive',
}

export function FilmStatusBadge({ status }: { status: string }) {
  const { t } = useTranslation()

  const labels: Record<string, string> = {
    ready:       t('admin.status_ready'),
    downloading: t('admin.status_downloading'),
    pending:     t('admin.status_pending'),
    error:       t('admin.status_error'),
  }

  const cls = STATUS_CLASSES[status] ?? 'bg-muted text-muted-foreground'

  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${cls}`}>
      {labels[status] ?? status}
    </span>
  )
}
