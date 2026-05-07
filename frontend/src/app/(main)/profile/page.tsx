'use client'
import React from 'react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useTranslation } from 'react-i18next'
import { CheckCircle, XCircle, Loader2 } from 'lucide-react'

const LANGUAGES = [
  { code: 'fr', label: '🇫🇷 Français' },
  { code: 'en', label: '🇬🇧 English' },
]

interface UserProfile {
  id: number
  username: string
  email: string
  first_name: string
  last_name: string
  avatar_url: string
}

type SaveStatus = 'idle' | 'loading' | 'success' | 'error'

export default function ProfilePage() {
  const { t, i18n } = useTranslation()
  const [isEditing, setIsEditing] = React.useState(false)
  const [profile, setProfile] = React.useState<UserProfile | null>(null)
  const [firstName, setFirstName] = React.useState('')
  const [lastName, setLastName] = React.useState('')
  const [selectedLang, setSelectedLang] = React.useState(i18n.language)
  const [profileImage, setProfileImage] = React.useState('')
  const [saveStatus, setSaveStatus] = React.useState<SaveStatus>('idle')
  const [errorMsg, setErrorMsg] = React.useState('')
  const fileInputRef = React.useRef<HTMLInputElement>(null)

  React.useEffect(() => {
    fetch('/api/v1/users/profile', { credentials: 'include' })
      .then((r) => r.json())
      .then(({ data }) => {
        const p: UserProfile = data.profile
        setProfile(p)
        setFirstName(p.first_name)
        setLastName(p.last_name)
        setProfileImage(p.avatar_url || `https://robohash.org/${p.id}.png?set=set1`)
      })
      .catch(() => {})
  }, [])

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (event) => setProfileImage(event.target?.result as string)
      reader.readAsDataURL(file)
    }
  }

  const handleSave = async () => {
    if (!profile) return
    setSaveStatus('loading')
    setErrorMsg('')
    try {
      const res = await fetch(`/api/v1/users/profile/${profile.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ first_name: firstName, last_name: lastName }),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        setErrorMsg(body.error ?? t('profile.save_error'))
        setSaveStatus('error')
        return
      }
      const { data } = await res.json()
      setProfile(data.profile)
      await i18n.changeLanguage(selectedLang)
      setSaveStatus('success')
      setIsEditing(false)
      setTimeout(() => setSaveStatus('idle'), 3000)
    } catch {
      setErrorMsg(t('profile.save_error'))
      setSaveStatus('error')
    }
  }

  const handleCancel = () => {
    if (profile) {
      setFirstName(profile.first_name)
      setLastName(profile.last_name)
      setProfileImage(profile.avatar_url || `https://robohash.org/${profile.id}.png?set=set1`)
    }
    setSelectedLang(i18n.language)
    setSaveStatus('idle')
    setIsEditing(false)
  }

  return (
    <div className="container mx-auto p-6 max-w-2xl">
      <div className="flex mb-6 items-center justify-between gap-4">
        <h1 className="text-3xl font-bold">{t('profile.title')}</h1>
        <div className="flex items-center gap-2">
          {isEditing ? (
            <>
              <Button variant="outline" onClick={handleCancel} disabled={saveStatus === 'loading'}>
                {t('profile.cancel')}
              </Button>
              <Button variant="default" onClick={handleSave} disabled={saveStatus === 'loading'}>
                {saveStatus === 'loading' ? (
                  <Loader2 className="size-4 animate-spin mr-1" />
                ) : null}
                {t('profile.save')}
              </Button>
            </>
          ) : (
            <Button variant="default" onClick={() => setIsEditing(true)}>
              {t('profile.edit')}
            </Button>
          )}
        </div>
      </div>

      {saveStatus === 'success' && (
        <div className="flex items-center gap-2 mb-4 text-sm text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 rounded-md px-4 py-2">
          <CheckCircle className="size-4 shrink-0" />
          {t('profile.save_success')}
        </div>
      )}

      {saveStatus === 'error' && (
        <div className="flex items-center gap-2 mb-4 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-md px-4 py-2">
          <XCircle className="size-4 shrink-0" />
          {errorMsg}
        </div>
      )}

      <Card className="card-glow">
        <CardHeader>
          <CardTitle className="text-lg">{t('profile.info')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-8 items-start">
            <div className="flex flex-col gap-4 flex-1">
              {profile?.username && (
                <div className="space-y-1.5">
                  <label className="text-sm font-medium text-foreground">{t('profile.username')}</label>
                  <Input value={profile.username} disabled />
                </div>
              )}

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">{t('Email')}</label>
                <Input value={profile?.email ?? ''} disabled />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">{t('Name')}</label>
                <Input
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                  disabled={!isEditing}
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">{t('First name')}</label>
                <Input
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  disabled={!isEditing}
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">{t('profile.language')}</label>
                <div className="flex gap-2">
                  {LANGUAGES.map((lang) => (
                    <button
                      key={lang.code}
                      type="button"
                      disabled={!isEditing}
                      onClick={() => isEditing && setSelectedLang(lang.code)}
                      className={`flex-1 rounded-md border px-3 py-2 text-sm transition-colors ${
                        selectedLang === lang.code
                          ? 'border-sidebar-primary bg-sidebar-primary/10 font-semibold text-sidebar-primary'
                          : 'border-border text-muted-foreground'
                      } disabled:opacity-50 disabled:cursor-not-allowed`}
                    >
                      {lang.label}
                    </button>
                  ))}
                </div>
              </div>
            </div>

            <div className="flex flex-col items-center gap-2 flex-shrink-0">
              <div
                className={`relative w-40 h-40 rounded-lg overflow-hidden border-2 border-border transition-colors ${isEditing ? 'cursor-pointer hover:border-primary' : ''}`}
                onClick={() => isEditing && fileInputRef.current?.click()}
              >
                {profileImage && (
                  <img src={profileImage} alt={t('profile.avatar_alt')} className="w-full h-full object-cover" />
                )}
                {isEditing && (
                  <div className="absolute inset-0 bg-background/60 flex items-center justify-center">
                    <span className="text-foreground text-sm font-medium">{t('profile.edit')}</span>
                  </div>
                )}
              </div>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleImageUpload}
                className="hidden"
              />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
