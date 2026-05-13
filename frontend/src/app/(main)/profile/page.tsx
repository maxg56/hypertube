'use client'
import React from 'react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useTranslation } from 'react-i18next'
import { CheckCircle, XCircle, Loader2, Globe, Lock, Heart } from 'lucide-react'
import { apiClient } from '@/lib/api'

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
  language: string
  is_public: boolean
  favorites_public: boolean
}

type SaveStatus = 'idle' | 'loading' | 'success' | 'error'

function avatarSrc(profile: UserProfile | null, localPreview: string): string {
  if (localPreview) return localPreview
  if (profile?.avatar_url) return profile.avatar_url
  if (profile?.id) return `https://robohash.org/${profile.id}.png?set=set1`
  return ''
}

export default function ProfilePage() {
  const { t, i18n } = useTranslation()
  const [isEditing, setIsEditing] = React.useState(false)
  const [profile, setProfile] = React.useState<UserProfile | null>(null)
  const [firstName, setFirstName] = React.useState('')
  const [lastName, setLastName] = React.useState('')
  const [selectedLang, setSelectedLang] = React.useState(i18n.language)
  const [isPublic, setIsPublic] = React.useState(true)
  const [favoritesPublic, setFavoritesPublic] = React.useState(true)
  const [pendingFile, setPendingFile] = React.useState<File | null>(null)
  const [localPreview, setLocalPreview] = React.useState('')
  const [saveStatus, setSaveStatus] = React.useState<SaveStatus>('idle')
  const [errorMsg, setErrorMsg] = React.useState('')
  const fileInputRef = React.useRef<HTMLInputElement>(null)

  React.useEffect(() => {
    apiClient.get<{ data: { profile: UserProfile } }>('/users/profile')
      .then(({ data }) => {
        const p: UserProfile = data.profile
        setProfile(p)
        setFirstName(p.first_name)
        setLastName(p.last_name)
        setIsPublic(p.is_public ?? true)
        setFavoritesPublic(p.favorites_public ?? true)
        if (p.language) {
          setSelectedLang(p.language)
          i18n.changeLanguage(p.language)
        }
      })
      .catch(() => {})
  }, [])

  const handleImagePick = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setPendingFile(file)
    const reader = new FileReader()
    reader.onload = (ev) => setLocalPreview(ev.target?.result as string)
    reader.readAsDataURL(file)
  }

  const handleSave = async () => {
    if (!profile) return
    setSaveStatus('loading')
    setErrorMsg('')
    try {
      if (pendingFile) {
        const form = new FormData()
        form.append('avatar', pendingFile)
        const avatarJson = await apiClient.post<{ data: { avatar_url: string } }>('/users/avatar', form)
        setProfile((prev) => prev ? { ...prev, avatar_url: avatarJson.data.avatar_url } : prev)
        setLocalPreview('')
        setPendingFile(null)
      }

      const profileJson = await apiClient.put<{ data: { profile: UserProfile } }>(
        `/users/profile/${profile.id}`,
        { first_name: firstName, last_name: lastName, language: selectedLang, is_public: isPublic, favorites_public: favoritesPublic },
      )
      setProfile(profileJson.data.profile)
      await i18n.changeLanguage(selectedLang)
      setSaveStatus('success')
      setIsEditing(false)
      setTimeout(() => setSaveStatus('idle'), 3000)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : ''
      setErrorMsg(message || t('profile.save_error'))
      setSaveStatus('error')
    }
  }

  const handleCancel = () => {
    if (profile) {
      setFirstName(profile.first_name)
      setLastName(profile.last_name)
      setIsPublic(profile.is_public ?? true)
      setFavoritesPublic(profile.favorites_public ?? true)
    }
    setLocalPreview('')
    setPendingFile(null)
    setSelectedLang(i18n.language)
    setSaveStatus('idle')
    setIsEditing(false)
  }

  const imgSrc = avatarSrc(profile, localPreview)

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
                {saveStatus === 'loading' && <Loader2 className="size-4 animate-spin mr-1" />}
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
                <Input value={lastName} onChange={(e) => setLastName(e.target.value)} disabled={!isEditing} />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">{t('First name')}</label>
                <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} disabled={!isEditing} />
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
                {imgSrc && (
                  <img src={imgSrc} alt={t('profile.avatar_alt')} className="w-full h-full object-cover" />
                )}
                {isEditing && (
                  <div className="absolute inset-0 bg-background/60 flex items-center justify-center">
                    <span className="text-foreground text-sm font-medium">{t('profile.change_avatar')}</span>
                  </div>
                )}
              </div>
              <input ref={fileInputRef} type="file" accept="image/*" onChange={handleImagePick} className="hidden" />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card className="card-glow">
        <CardHeader>
          <CardTitle className="text-lg">{t('profile.privacy_title')}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <PrivacyToggle
            icon={isPublic ? <Globe className="size-4 text-sidebar-primary" /> : <Lock className="size-4 text-muted-foreground" />}
            label={t('profile.privacy_account_public')}
            hint={t('profile.privacy_account_public_hint')}
            checked={isPublic}
            disabled={!isEditing}
            onChange={setIsPublic}
          />
          <PrivacyToggle
            icon={<Heart className={`size-4 ${favoritesPublic ? 'text-destructive' : 'text-muted-foreground'}`} />}
            label={t('profile.privacy_favorites_public')}
            hint={t('profile.privacy_favorites_public_hint')}
            checked={favoritesPublic}
            disabled={!isEditing || !isPublic}
            onChange={setFavoritesPublic}
          />
        </CardContent>
      </Card>
    </div>
  )
}

function PrivacyToggle({
  icon, label, hint, checked, disabled, onChange,
}: {
  icon: React.ReactNode
  label: string
  hint: string
  checked: boolean
  disabled: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <div className={`flex items-center justify-between gap-4 ${disabled ? 'opacity-50' : ''}`}>
      <div className="flex items-center gap-3 min-w-0">
        {icon}
        <div className="min-w-0">
          <p className="text-sm font-medium">{label}</p>
          <p className="text-xs text-muted-foreground">{hint}</p>
        </div>
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        disabled={disabled}
        onClick={() => !disabled && onChange(!checked)}
        className={`relative inline-flex h-6 w-11 shrink-0 rounded-full border-2 border-transparent transition-colors focus-visible:outline-none disabled:cursor-not-allowed ${
          checked ? 'bg-sidebar-primary' : 'bg-muted'
        }`}
      >
        <span
          className={`pointer-events-none block h-5 w-5 rounded-full bg-white shadow-lg transition-transform ${
            checked ? 'translate-x-5' : 'translate-x-0'
          }`}
        />
      </button>
    </div>
  )
}
