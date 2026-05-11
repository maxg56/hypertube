'use client'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Loader2, Trash2, RefreshCw } from 'lucide-react'
import { FilmStatusBadge } from './FilmStatusBadge'
import type { AdminTorrent } from '../hooks/useAdminFilms'

function formatBytes(bytes: number): string {
  if (bytes === 0) return '—'
  const units = ['B', 'KB', 'MB', 'GB']
  let v = bytes
  let i = 0
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++ }
  return `${v.toFixed(1)} ${units[i]}`
}

interface FilmRowProps {
  torrent: AdminTorrent
  isActing: boolean
  actionType: 'delete' | 'redownload' | null
  onDelete: () => void
  onReDownload: () => void
}

export function FilmRow({ torrent, isActing, actionType, onDelete, onReDownload }: FilmRowProps) {
  const { t } = useTranslation()

  return (
    <tr className="border-b last:border-0 hover:bg-muted/40 transition-colors">
      <td className="py-2.5 pl-6 pr-4">
        <span className="font-mono text-xs text-muted-foreground" title={torrent.info_hash}>
          {torrent.info_hash.slice(0, 16)}…
        </span>
        {torrent.status === 'downloading' && (
          <div className="mt-1 flex items-center gap-2">
            <div className="h-1 w-24 rounded-full bg-muted overflow-hidden">
              <div
                className="h-full bg-sidebar-primary transition-all"
                style={{ width: `${Math.min(100, torrent.progress)}%` }}
              />
            </div>
            <span className="text-xs text-muted-foreground">{torrent.progress.toFixed(0)}%</span>
          </div>
        )}
      </td>

      <td className="py-2.5 pr-4">
        {torrent.quality ? (
          <span className="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium bg-muted text-muted-foreground uppercase tracking-wide">
            {torrent.quality}
          </span>
        ) : (
          <span className="text-muted-foreground text-xs">—</span>
        )}
      </td>

      <td className="py-2.5 pr-4">
        <FilmStatusBadge status={torrent.status} />
      </td>

      <td className="py-2.5 pr-4 text-right tabular-nums text-muted-foreground">
        {formatBytes(torrent.file_size)}
      </td>

      <td className="py-2.5 text-right">
        <div className="flex items-center justify-end gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={isActing}
            onClick={onReDownload}
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
            onClick={onDelete}
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
}
