'use client'
import Link from 'next/link'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import type { UserResult } from '@/hooks/useUserSearch'

export function UserCard({ user }: { user: UserResult }) {
  const avatarSrc = user.avatar_url || `https://robohash.org/${user.id}.png?set=set1`
  const initials = ((user.first_name?.[0] ?? '') + (user.last_name?.[0] ?? '')).toUpperCase() || user.username?.[0]?.toUpperCase() || '?'
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ')

  return (
    <Link
      href={`/users/${user.id}`}
      className="flex flex-col items-center gap-3 p-4 rounded-xl border border-border bg-card hover:bg-muted transition-colors text-center group"
    >
      <Avatar className="size-20 rounded-full ring-2 ring-border group-hover:ring-sidebar-primary/50 transition-all">
        <AvatarImage src={avatarSrc} alt={user.username} className="object-cover" />
        <AvatarFallback className="text-lg font-semibold">{initials}</AvatarFallback>
      </Avatar>
      <div className="min-w-0 w-full">
        <p className="font-semibold truncate text-sm">{user.username}</p>
        {fullName && <p className="text-xs text-muted-foreground truncate">{fullName}</p>}
      </div>
    </Link>
  )
}
