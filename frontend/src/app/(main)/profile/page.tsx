'use client'
import React from 'react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ThemeToggle } from '@/components/page/ThemeToggle'
import { useTranslation } from 'react-i18next'

export default function ProfilePage() {
  const [isEditing, setIsEditing] = React.useState(false)
  const [profileImage, setProfileImage] = React.useState('https://robohash.org/1.png?set=set1')
  const fileInputRef = React.useRef<HTMLInputElement>(null)
  const { t } = useTranslation()

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (event) => {
        setProfileImage(event.target?.result as string)
      }
      reader.readAsDataURL(file)
    }
  }

  return (
    <div className="container mx-auto p-6 max-w-2xl">
      <div className="flex mb-6 items-center justify-between gap-4">
        <h1 className="text-3xl font-bold">{t('Profile')}</h1>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <Button variant="default" onClick={() => setIsEditing(!isEditing)}>
            {isEditing ? t('profile.save') : t('profile.edit')}
          </Button>
        </div>
      </div>

      <Card className="card-glow">
        <CardHeader>
          <CardTitle className="text-lg">{t('profile.info')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-8 items-start">
            <div className="flex flex-col gap-4 flex-1">
              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">{t('Email')}</label>
                <Input
                  placeholder="Email"
                  defaultValue="john.doe@example.com"
                  disabled={!isEditing}
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">{t('Name')}</label>
                <Input
                  placeholder="Nom"
                  defaultValue="Doe"
                  disabled={!isEditing}
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">{t('First name')}</label>
                <Input
                  placeholder="Prénom"
                  defaultValue="John"
                  disabled={!isEditing}
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-sm font-medium text-foreground">{t('Password')}</label>
                <Input
                  type="password"
                  placeholder="••••••••"
                  disabled={!isEditing}
                />
              </div>
            </div>

            <div className="flex flex-col items-center gap-2 flex-shrink-0">
              <div
                className={`relative w-40 h-40 rounded-lg overflow-hidden border-2 border-border transition-colors ${isEditing ? 'cursor-pointer hover:border-primary' : ''}`}
                onClick={() => isEditing && fileInputRef.current?.click()}
              >
                <img
                  src={profileImage}
                  alt="Profil"
                  className="w-full h-full object-cover"
                />
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
