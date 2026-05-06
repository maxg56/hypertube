'use client'

import { useActionState } from 'react'
import Link from 'next/link'
import { useTranslation } from 'react-i18next'
import { register } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function RegisterForm() {
  const { t } = useTranslation()
  const [state, action, pending] = useActionState(register, undefined)

  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">{t('auth.register_title')}</h2>
        <p className="text-sm text-muted-foreground mt-1">{t('auth.register_subtitle')}</p>
      </div>

      <form action={action} className="space-y-4">
        {state?.message && (
          <p className="text-sm text-destructive text-center">{state.message}</p>
        )}

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <Label htmlFor="first_name">{t('auth.register_first_name')}</Label>
            <Input
              id="first_name"
              name="first_name"
              type="text"
              autoComplete="given-name"
              placeholder={t('auth.register_first_name_placeholder')}
              required
            />
            {state?.errors?.first_name && (
              <p className="text-xs text-destructive">{state.errors.first_name[0]}</p>
            )}
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="last_name">{t('auth.register_last_name')}</Label>
            <Input
              id="last_name"
              name="last_name"
              type="text"
              autoComplete="family-name"
              placeholder={t('auth.register_last_name_placeholder')}
              required
            />
            {state?.errors?.last_name && (
              <p className="text-xs text-destructive">{state.errors.last_name[0]}</p>
            )}
          </div>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="username">{t('auth.register_username')}</Label>
          <Input
            id="username"
            name="username"
            type="text"
            autoComplete="username"
            placeholder={t('auth.register_username_placeholder')}
            required
          />
          {state?.errors?.username && (
            <p className="text-xs text-destructive">{state.errors.username[0]}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="email">{t('auth.register_email')}</Label>
          <Input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            placeholder={t('auth.register_email_placeholder')}
            required
          />
          {state?.errors?.email && (
            <p className="text-xs text-destructive">{state.errors.email[0]}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="password">{t('auth.register_password')}</Label>
          <Input
            id="password"
            name="password"
            type="password"
            autoComplete="new-password"
            placeholder={t('auth.register_password_placeholder')}
            required
          />
          {state?.errors?.password && (
            <ul className="text-xs text-destructive space-y-0.5">
              {state.errors.password.map((e) => (
                <li key={e}>• {e}</li>
              ))}
            </ul>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={pending}>
          {pending ? t('auth.register_submitting') : t('auth.register_submit')}
        </Button>

        <p className="text-center text-sm text-muted-foreground">
          {t('auth.register_already_account')}{' '}
          <Link href="/login" className="text-primary hover:underline">
            {t('auth.register_login_link')}
          </Link>
        </p>
      </form>
    </>
  )
}
