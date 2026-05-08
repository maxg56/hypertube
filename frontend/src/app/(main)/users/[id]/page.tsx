'use client'
import React from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { Heart } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import { useTranslation } from 'react-i18next'
import { MovieCard } from '@/components/page/MovieCard'
import type { Movie } from '@/hooks/useMovies'

interface PublicProfile {
  id: number
  username: string
  first_name: string
  last_name: string
  avatar_url: string
}

interface UserComment {
  id: number
  content: string
  created_at: string
  movie_id: number
  tmdb_id: number
  title: string
}

interface FavoriteMovie {
  tmdb_id: number
  title: string
  poster_url: string
  rating: number
  release_date: string
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
  } catch {
    return iso
  }
}

function toMovie(f: FavoriteMovie): Movie {
  return {
    id: f.tmdb_id,
    imdb_id: '',
    title: f.title,
    year: f.release_date?.slice(0, 4) ?? '',
    rating: f.rating,
    poster_url: f.poster_url,
    genres: [],
  }
}

export default function PublicProfilePage() {
  const { id } = useParams<{ id: string }>()
  const { t } = useTranslation()
  const [profile, setProfile] = React.useState<PublicProfile | null>(null)
  const [comments, setComments] = React.useState<UserComment[]>([])
  const [favorites, setFavorites] = React.useState<FavoriteMovie[]>([])
  const [notFound, setNotFound] = React.useState(false)

  React.useEffect(() => {
    fetch(`/api/v1/users/profile/${id}`, { credentials: 'include' })
      .then((r) => {
        if (r.status === 404) { setNotFound(true); return null }
        return r.json()
      })
      .then((body) => { if (body) setProfile(body.data.profile) })
      .catch(() => setNotFound(true))

    fetch(`/api/v1/comments/user/${id}`, { credentials: 'include' })
      .then((r) => r.json())
      .then(({ data }) => setComments(data ?? []))
      .catch(() => {})

    fetch(`/api/v1/users/${id}/favorites?limit=12`, { credentials: 'include' })
      .then((r) => r.json())
      .then(({ data }) => setFavorites(data?.favorites ?? []))
      .catch(() => {})
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
    <div className="container mx-auto p-6 max-w-4xl flex flex-col gap-6">
      <h1 className="text-3xl font-bold">{t('profile.public_title')}</h1>

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

      {favorites.length > 0 && (
        <Card className="card-glow">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Heart className="size-4 text-destructive fill-current" />
              {t('favorites.title')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 gap-3">
              {favorites.map((f) => (
                <MovieCard key={f.tmdb_id} movie={toMovie(f)} watched={false} />
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      <Card className="card-glow">
        <CardHeader>
          <CardTitle className="text-lg">{t('profile.comments_title')}</CardTitle>
        </CardHeader>
        <CardContent>
          {comments.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t('profile.no_comments')}</p>
          ) : (
            <div className="flex flex-col divide-y divide-border">
              {comments.map((c) => (
                <div key={c.id} className="py-3 first:pt-0 last:pb-0 flex flex-col gap-1">
                  <div className="flex items-center justify-between gap-2">
                    <Link
                      href={`/movies/${c.tmdb_id}`}
                      className="text-sm font-medium text-sidebar-primary hover:underline truncate"
                    >
                      {c.title}
                    </Link>
                    <span className="text-xs text-muted-foreground shrink-0">{formatDate(c.created_at)}</span>
                  </div>
                  <p className="text-sm text-muted-foreground break-words">{c.content}</p>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
