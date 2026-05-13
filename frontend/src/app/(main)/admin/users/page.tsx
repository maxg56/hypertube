'use client'
import React from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Loader2, Pencil, Shield, Trash2 } from 'lucide-react'
import { apiClient } from '@/lib/api'

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

function ConfirmDialog({
  title, body, cancelLabel, confirmLabel, onCancel, onConfirm, danger = true,
}: {
  title: string
  body: string
  cancelLabel: string
  confirmLabel: string
  onCancel: () => void
  onConfirm: () => void
  danger?: boolean
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
      <div className="bg-card border border-border rounded-xl shadow-lg p-6 max-w-sm w-full mx-4">
        <h2 className="text-base font-semibold mb-2">{title}</h2>
        <p className="text-sm text-muted-foreground mb-6">{body}</p>
        <div className="flex gap-3 justify-end">
          <Button variant="outline" size="sm" onClick={onCancel}>{cancelLabel}</Button>
          <Button variant={danger ? 'destructive' : 'default'} size="sm" onClick={onConfirm}>{confirmLabel}</Button>
        </div>
      </div>
    </div>
  )
}

function RenameDialog({
  user, value, onChange, onCancel, onConfirm,
}: {
  user: AdminUser
  value: string
  onChange: (v: string) => void
  onCancel: () => void
  onConfirm: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
      <div className="bg-card border border-border rounded-xl shadow-lg p-6 max-w-sm w-full mx-4">
        <h2 className="text-base font-semibold mb-4">{t('admin.rename_title')} — {user.username}</h2>
        <label className="block text-sm font-medium mb-1">{t('admin.rename_label')}</label>
        <input
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 mb-6"
          placeholder={t('admin.rename_placeholder')}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') onConfirm() }}
          autoFocus
        />
        <div className="flex gap-3 justify-end">
          <Button variant="outline" size="sm" onClick={onCancel}>{t('admin.rename_cancel')}</Button>
          <Button size="sm" disabled={!value.trim()} onClick={onConfirm}>{t('admin.rename_submit')}</Button>
        </div>
      </div>
    </div>
  )
}

export default function AdminUsersPage() {
  const { t } = useTranslation()
  const [users, setUsers] = React.useState<AdminUser[]>([])
  const [total, setTotal] = React.useState(0)
  const [offset, setOffset] = React.useState(0)
  const [loading, setLoading] = React.useState(true)
  const [currentUserId, setCurrentUserId] = React.useState<number | null>(null)
  const [pendingPromote, setPendingPromote] = React.useState<AdminUser | null>(null)
  const [pendingDeleteUser, setPendingDeleteUser] = React.useState<AdminUser | null>(null)
  const [pendingRename, setPendingRename] = React.useState<AdminUser | null>(null)
  const [newUsername, setNewUsername] = React.useState('')
  const [acting, setActing] = React.useState<number | null>(null)
  const limit = 20

  React.useEffect(() => {
    apiClient.get<{ data: { id: number } }>('/users/profile')
      .then(json => { if (json.data?.id) setCurrentUserId(json.data.id) })
      .catch(() => {})
  }, [])

  const fetchUsers = React.useCallback(async (off: number) => {
    setLoading(true)
    try {
      const json = await apiClient.get<{ data: { users: AdminUser[]; pagination: { total: number } } }>(
        `/admin/users?limit=${limit}&offset=${off}`,
      )
      setUsers(json.data?.users ?? [])
      setTotal(json.data?.pagination?.total ?? 0)
    } catch {
      // keep previous state on error
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchUsers(0) }, [fetchUsers])

  const goTo = (off: number) => {
    setOffset(off)
    fetchUsers(off)
  }

  const promoteUser = async (user: AdminUser) => {
    setActing(user.id)
    const newRole = user.role === 'admin' ? 'user' : 'admin'
    try {
      await apiClient.put(`/admin/users/${user.id}/role`, { role: newRole })
      setUsers(prev => prev.map(u => u.id === user.id ? { ...u, role: newRole } : u))
    } catch {
      // keep previous state on error
    } finally {
      setActing(null)
    }
  }

  const deleteUser = async (user: AdminUser) => {
    setActing(user.id)
    try {
      await apiClient.delete(`/admin/users/${user.id}`)
      setUsers(prev => prev.filter(u => u.id !== user.id))
      setTotal(prev => prev - 1)
    } catch {
      // keep previous state on error
    } finally {
      setActing(null)
    }
  }

  const renameUser = async (user: AdminUser, username: string) => {
    setActing(user.id)
    try {
      await apiClient.put(`/admin/users/${user.id}/username`, { username: username.trim() })
      setUsers(prev => prev.map(u => u.id === user.id ? { ...u, username: username.trim() } : u))
    } catch {
      // keep previous state on error
    } finally {
      setActing(null)
    }
  }

  return (
    <>
      {pendingPromote && (
        <ConfirmDialog
          title={pendingPromote.role === 'admin' ? t('admin.confirm_demote_title') : t('admin.confirm_promote_title')}
          body={pendingPromote.role === 'admin' ? t('admin.confirm_demote_body') : t('admin.confirm_promote_body')}
          cancelLabel={t('admin.confirm_promote_cancel')}
          confirmLabel={pendingPromote.role === 'admin' ? t('admin.confirm_demote_confirm') : t('admin.confirm_promote_confirm')}
          danger={false}
          onCancel={() => setPendingPromote(null)}
          onConfirm={async () => {
            const u = pendingPromote
            setPendingPromote(null)
            await promoteUser(u)
          }}
        />
      )}

      {pendingDeleteUser && (
        <ConfirmDialog
          title={t('admin.confirm_delete_user_title')}
          body={t('admin.confirm_delete_user_body')}
          cancelLabel={t('admin.confirm_delete_user_cancel')}
          confirmLabel={t('admin.confirm_delete_user_confirm')}
          onCancel={() => setPendingDeleteUser(null)}
          onConfirm={async () => {
            const u = pendingDeleteUser
            setPendingDeleteUser(null)
            await deleteUser(u)
          }}
        />
      )}

      {pendingRename && (
        <RenameDialog
          user={pendingRename}
          value={newUsername}
          onChange={setNewUsername}
          onCancel={() => { setPendingRename(null); setNewUsername('') }}
          onConfirm={async () => {
            if (!newUsername.trim()) return
            const u = pendingRename
            const name = newUsername
            setPendingRename(null)
            setNewUsername('')
            await renameUser(u, name)
          }}
        />
      )}

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
                      <th className="text-right py-2 pr-4 font-medium">{t('admin.col_downloaded')}</th>
                      <th className="text-right py-2 font-medium">{t('admin.col_actions')}</th>
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
                        <td className="py-2.5 pr-4 text-right tabular-nums">{u.films_downloaded}</td>
                        <td className="py-2.5 text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button
                              variant="ghost"
                              size="sm"
                              disabled={acting === u.id}
                              onClick={() => { setNewUsername(u.username); setPendingRename(u) }}
                              title={t('admin.action_rename')}
                            >
                              <Pencil className="size-3.5" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              disabled={acting === u.id || u.id === currentUserId}
                              onClick={() => setPendingPromote(u)}
                              title={u.role === 'admin' ? t('admin.action_demote') : t('admin.action_promote')}
                              className={u.role === 'admin' ? 'text-sidebar-primary' : ''}
                            >
                              {acting === u.id
                                ? <Loader2 className="size-3.5 animate-spin" />
                                : <Shield className="size-3.5" />
                              }
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              disabled={acting === u.id || u.id === currentUserId}
                              onClick={() => setPendingDeleteUser(u)}
                              title={t('admin.action_delete_user')}
                              className="text-destructive hover:text-destructive"
                            >
                              <Trash2 className="size-3.5" />
                            </Button>
                          </div>
                        </td>
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
    </>
  )
}
