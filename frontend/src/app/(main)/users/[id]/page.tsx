'use client'
import React from 'react'
import { useParams } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import { useTranslation } from 'react-i18next'

interface PublicProfile {
  id: number
  username: string
  first_name: string
  last_name: string
  avatar_url: string
}

export default function PublicProfilePage() {
  const { id } = useParams<{ id: string }>()
  const { t } = useTranslation()
  const [profile, setProfile] = React.useState<PublicProfile | null>(null)
  const [notFound, setNotFound] = React.useState(false)

  React.useEffect(() => {
    fetch(`/api/v1/users/profile/${id}`, { credentials: 'include' })
      .then((r) => {
        if (r.status === 404) { setNotFound(true); return null }
        return r.json()
      })
      .then((body) => {
        if (body) setProfile(body.data.profile)
      })
      .catch(() => setNotFound(true))
  }, [id])

  if (notFound) {
    return (
      <div className="container mx-auto p-6 max-w-2xl text-center text-muted-foreground">
        {t('profile.user_not_found')}
      </div>
    )
  }

  if (!profile) {
    return (
      <div className="container mx-auto p-6 max-w-2xl text-center text-muted-foreground">
        {t('profile.loading')}
      </div>
    )
  }

  const avatarSrc = profile.avatar_url || `https://robohash.org/${profile.id}.png?set=set1`
  const initials = (profile.first_name?.[0] ?? '') + (profile.last_name?.[0] ?? '') || profile.username?.[0]?.toUpperCase()

  return (
    <div className="container mx-auto p-6 max-w-2xl">
      <h1 className="text-3xl font-bold mb-6">{t('profile.public_title')}</h1>
      <Card className="card-glow">
        <CardHeader>
          <CardTitle className="text-lg">{t('profile.info')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-8 items-center">
            <Avatar className="size-28 rounded-lg">
              <AvatarImage src={avatarSrc} alt={profile.username} className="object-cover" />
              <AvatarFallback className="rounded-lg text-2xl">{initials}</AvatarFallback>
            </Avatar>
            <div className="flex flex-col gap-3">
              <div>
                <p className="text-xs text-muted-foreground mb-0.5">{t('profile.username')}</p>
                <p className="font-semibold text-lg">{profile.username}</p>
              </div>
              {(profile.first_name || profile.last_name) && (
                <div>
                  <p className="text-xs text-muted-foreground mb-0.5">{t('profile.full_name')}</p>
                  <p className="font-medium">{[profile.first_name, profile.last_name].filter(Boolean).join(' ')}</p>
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
