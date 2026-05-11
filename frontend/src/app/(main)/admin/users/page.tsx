'use client'
import React from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Loader2 } from 'lucide-react'

interface AdminUser {
  id: number
  username: string
  email: string
  first_name: string
  last_name: string
  avatar_url: string
  role: string
  email_verified: boolean
  created_at: string
  films_watched: number
  films_downloaded: number
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
  } catch {
    return iso
  }
}

export default function AdminUsersPage() {
  const { t } = useTranslation()
  const [users, setUsers] = React.useState<AdminUser[]>([])
  const [total, setTotal] = React.useState(0)
  const [offset, setOffset] = React.useState(0)
  const [loading, setLoading] = React.useState(true)
  const limit = 20

  const fetchUsers = React.useCallback(async (off: number) => {
    setLoading(true)
    try {
      const res = await fetch(`/api/v1/admin/users?limit=${limit}&offset=${off}`, { credentials: 'include' })
      if (!res.ok) return
      const json = await res.json()
      setUsers(json.data?.users ?? [])
      setTotal(json.data?.pagination?.total ?? 0)
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchUsers(0) }, [fetchUsers])

  const goTo = (off: number) => {
    setOffset(off)
    fetchUsers(off)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('admin.users_title')}</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="size-6 animate-spin text-muted-foreground" />
          </div>
        ) : users.length === 0 ? (
          <p className="text-muted-foreground text-sm py-8 text-center">{t('admin.no_users')}</p>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-muted-foreground">
                    <th className="text-left py-2 pr-4 font-medium">{t('admin.col_username')}</th>
                    <th className="text-left py-2 pr-4 font-medium">{t('admin.col_email')}</th>
                    <th className="text-left py-2 pr-4 font-medium">{t('admin.col_role')}</th>
                    <th className="text-left py-2 pr-4 font-medium">{t('admin.col_registered')}</th>
                    <th className="text-right py-2 pr-4 font-medium">{t('admin.col_watched')}</th>
                    <th className="text-right py-2 font-medium">{t('admin.col_downloaded')}</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((u) => (
                    <tr key={u.id} className="border-b last:border-0 hover:bg-muted/40 transition-colors">
                      <td className="py-2.5 pr-4">
                        <div className="flex items-center gap-2">
                          <img
                            src={u.avatar_url || `https://robohash.org/${u.id}.png?set=set1`}
                            alt={u.username}
                            className="size-7 rounded-full object-cover shrink-0"
                          />
                          <span className="font-medium">{u.username}</span>
                        </div>
                      </td>
                      <td className="py-2.5 pr-4 text-muted-foreground">{u.email}</td>
                      <td className="py-2.5 pr-4">
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                          u.role === 'admin'
                            ? 'bg-sidebar-primary/15 text-sidebar-primary'
                            : 'bg-muted text-muted-foreground'
                        }`}>
                          {u.role}
                        </span>
                      </td>
                      <td className="py-2.5 pr-4 text-muted-foreground">{formatDate(u.created_at)}</td>
                      <td className="py-2.5 pr-4 text-right tabular-nums">{u.films_watched}</td>
                      <td className="py-2.5 text-right tabular-nums">{u.films_downloaded}</td>
                    </tr>
                  ))}
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
  )
}
