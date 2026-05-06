'use client'

import { useActionState } from 'react'
import Link from 'next/link'
import { useTranslation } from 'react-i18next'
import { resetPassword } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function ResetPasswordForm({ token }: Readonly<{ token: string }>) {
  const { t } = useTranslation()
  const [state, action, pending] = useActionState(resetPassword, undefined)

  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">{t('auth.reset_password_title')}</h2>
        <p className="text-sm text-muted-foreground mt-1">{t('auth.reset_password_subtitle')}</p>
      </div>

      <form action={action} className="space-y-4">
        <input type="hidden" name="token" value={token} />

        {state?.message && (
          <p className="text-sm text-destructive text-center">{state.message}</p>
        )}

        <div className="space-y-1.5">
          <Label htmlFor="new_password">{t('auth.reset_password_new_label')}</Label>
          <Input
            id="new_password"
            name="new_password"
            type="password"
            autoComplete="new-password"
            placeholder={t('auth.reset_password_placeholder')}
            required
          />
          {state?.errors?.new_password && (
            <p className="text-xs text-destructive">{state.errors.new_password[0]}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="confirm_password">{t('auth.reset_password_confirm_label')}</Label>
          <Input
            id="confirm_password"
            name="confirm_password"
            type="password"
            autoComplete="new-password"
            placeholder={t('auth.reset_password_placeholder')}
            required
          />
          {state?.errors?.confirm_password && (
            <p className="text-xs text-destructive">{state.errors.confirm_password[0]}</p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={pending}>
          {pending ? t('auth.reset_password_submitting') : t('auth.reset_password_submit')}
        </Button>

        <p className="text-center text-sm text-muted-foreground">
          <Link href="/login" className="text-primary hover:underline">
            {t('auth.reset_password_back')}
          </Link>
        </p>
      </form>
    </>
  )
}
