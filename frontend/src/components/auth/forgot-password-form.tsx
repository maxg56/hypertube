'use client'

import { useActionState } from 'react'
import Link from 'next/link'
import { useTranslation } from 'react-i18next'
import { forgotPassword } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function ForgotPasswordForm() {
  const { t } = useTranslation()
  const [state, action, pending] = useActionState(forgotPassword, undefined)

  if (state?.success) {
    return (
      <div className="space-y-4 text-center">
        <p className="text-sm text-green-600 dark:text-green-400">{state.success}</p>
        <Link href="/login" className="text-sm text-primary hover:underline">
          {t('auth.forgot_password_back')}
        </Link>
      </div>
    )
  }

  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">{t('auth.forgot_password_title')}</h2>
        <p className="text-sm text-muted-foreground mt-1">{t('auth.forgot_password_subtitle')}</p>
      </div>

      <form action={action} className="space-y-4">
        {state?.message && (
          <p className="text-sm text-destructive text-center">{state.message}</p>
        )}

        <div className="space-y-1.5">
          <Label htmlFor="email">{t('auth.forgot_password_email')}</Label>
          <Input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            placeholder={t('auth.forgot_password_email_placeholder')}
            required
          />
          {state?.errors?.email && (
            <p className="text-xs text-destructive">{state.errors.email[0]}</p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={pending}>
          {pending ? t('auth.forgot_password_submitting') : t('auth.forgot_password_submit')}
        </Button>

        <p className="text-center text-sm text-muted-foreground">
          <Link href="/login" className="text-primary hover:underline">
            {t('auth.forgot_password_back')}
          </Link>
        </p>
      </form>
    </>
  )
}
