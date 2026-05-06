'use client'

import { useActionState, useEffect } from 'react'
import Link from 'next/link'
import { useTranslation } from 'react-i18next'
import { login } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuthContext } from '@/context/auth-context'

const API_URL = (process.env.NEXT_PUBLIC_API_URL ?? 'https://localhost:8443/api/v1').replace(/\/api\/v1\/?$/, '')

function Icon42() {
  return (
    <svg viewBox="0 0 24 24" className="size-4 mr-2" fill="currentColor" aria-hidden="true">
      <path d="M0 16.5V12l9-9h4.5L4.5 12H9v-4.5L18 0h4.5L13.5 9H18V0h4.5v13.5H9V18H4.5v-4.5H0v3H4.5V18H0ZM18 13.5v-4.5l-4.5 4.5H18Z" />
    </svg>
  )
}

function IconGitHub() {
  return (
    <svg viewBox="0 0 24 24" className="size-4 mr-2" fill="currentColor" aria-hidden="true">
      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61-.546-1.385-1.335-1.755-1.335-1.755-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23a11.5 11.5 0 0 1 3-.405c1.02.005 2.045.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 21.795 24 17.295 24 12c0-6.63-5.37-12-12-12Z" />
    </svg>
  )
}

export default function LoginForm({
  onSuccess,
  callbackUrl,
}: Readonly<{
  onSuccess?: () => void
  callbackUrl?: string
}>) {
  const { t } = useTranslation()
  const [state, action, pending] = useActionState(login, undefined)
  const { triggerLoginSuccess } = useAuthContext()

  useEffect(() => {
    if (state && !state.message && !state.errors) {
      onSuccess?.()
      triggerLoginSuccess()
    }
  }, [state, onSuccess, triggerLoginSuccess])

  const safeCallback = callbackUrl?.startsWith('/') ? callbackUrl : '/'

  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">{t('auth.login_title')}</h2>
        <p className="text-sm text-muted-foreground mt-1">{t('auth.login_subtitle')}</p>
      </div>

      <form action={action} className="space-y-4">
        <input type="hidden" name="callbackUrl" value={safeCallback} />

        {state?.message && (
          <p className="text-sm text-destructive text-center">{state.message}</p>
        )}

        <div className="space-y-1.5">
          <Label htmlFor="login">{t('auth.login_identifier_label')}</Label>
          <Input
            id="login"
            name="login"
            type="text"
            autoComplete="username"
            placeholder={t('auth.login_identifier_placeholder')}
            required
          />
          {state?.errors?.login && (
            <p className="text-xs text-destructive">{state.errors.login[0]}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label htmlFor="password">{t('auth.login_password_label')}</Label>
            <Link
              href="/forgot-password"
              className="text-xs text-muted-foreground hover:text-primary transition-colors"
            >
              {t('auth.login_forgot_password')}
            </Link>
          </div>
          <Input
            id="password"
            name="password"
            type="password"
            autoComplete="current-password"
            placeholder="••••••••"
            required
          />
          {state?.errors?.password && (
            <p className="text-xs text-destructive">{state.errors.password[0]}</p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={pending}>
          {pending ? t('auth.login_submitting') : t('auth.login_submit')}
        </Button>

        <div className="relative my-2">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-card px-2 text-muted-foreground">{t('auth.login_or')}</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <a href={`${API_URL}/api/v1/auth/oauth/42?callbackUrl=${encodeURIComponent(safeCallback)}`}>
            <Button variant="outline" className="w-full" type="button">
              <Icon42 />
              42 School
            </Button>
          </a>
          <a href={`${API_URL}/api/v1/auth/oauth/github?callbackUrl=${encodeURIComponent(safeCallback)}`}>
            <Button variant="outline" className="w-full" type="button">
              <IconGitHub />
              GitHub
            </Button>
          </a>
        </div>

        <p className="text-center text-sm text-muted-foreground">
          {t('auth.login_no_account')}{' '}
          <Link href="/register" className="text-primary hover:underline">
            {t('auth.login_register_link')}
          </Link>
        </p>
      </form>
    </>
  )
}
