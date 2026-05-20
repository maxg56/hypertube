'use client'

import { useActionState } from 'react'
import Link from 'next/link'
import { useTranslation } from 'react-i18next'
import { register } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/ui/password-input'
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
          <p role="alert" className="text-sm text-destructive text-center">{state.message}</p>
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
              aria-invalid={!!state?.errors?.first_name}
              aria-describedby={state?.errors?.first_name ? 'err-first_name' : undefined}
              required
            />
            {state?.errors?.first_name && (
              <p id="err-first_name" role="alert" className="text-xs text-destructive">{state.errors.first_name[0]}</p>
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
              aria-invalid={!!state?.errors?.last_name}
              aria-describedby={state?.errors?.last_name ? 'err-last_name' : undefined}
              required
            />
            {state?.errors?.last_name && (
              <p id="err-last_name" role="alert" className="text-xs text-destructive">{state.errors.last_name[0]}</p>
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
            aria-invalid={!!state?.errors?.username}
            aria-describedby={state?.errors?.username ? 'err-username' : undefined}
            required
          />
          {state?.errors?.username && (
            <p id="err-username" role="alert" className="text-xs text-destructive">{state.errors.username[0]}</p>
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
            aria-invalid={!!state?.errors?.email}
            aria-describedby={state?.errors?.email ? 'err-email' : undefined}
            required
          />
          {state?.errors?.email && (
            <p id="err-email" role="alert" className="text-xs text-destructive">{state.errors.email[0]}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="password">{t('auth.register_password')}</Label>
          <PasswordInput
            id="password"
            name="password"
            autoComplete="new-password"
            placeholder={t('auth.register_password_placeholder')}
            aria-invalid={!!state?.errors?.password}
            aria-describedby="hint-password err-password"
          />
          <p id="hint-password" className="text-xs text-muted-foreground">
            {t('auth.register_password_hint')}
          </p>
          {state?.errors?.password && (
            <ul id="err-password" role="alert" className="text-xs text-destructive space-y-0.5">
              {state.errors.password.map((e) => (
                <li key={e}>• {e}</li>
              ))}
            </ul>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="confirm_password">{t('auth.register_confirm_password')}</Label>
          <Input
            id="confirm_password"
            name="confirm_password"
            type="password"
            autoComplete="new-password"
            placeholder={t('auth.register_confirm_password_placeholder')}
            required
          />
          {state?.errors?.confirm_password && (
            <p className="text-xs text-destructive">{state.errors.confirm_password[0]}</p>
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
