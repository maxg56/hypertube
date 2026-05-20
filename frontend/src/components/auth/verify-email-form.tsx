'use client'

import { useActionState } from 'react'
import Link from 'next/link'
import { useTranslation } from 'react-i18next'
import { verifyEmail, sendEmailVerification, confirmEmailChange } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function VerifyEmailForm({
  defaultEmail,
  mode,
}: Readonly<{ defaultEmail?: string; mode?: string }>) {
  const { t } = useTranslation()
  const isChangeMode = mode === 'change'

  const verifyAction = isChangeMode ? confirmEmailChange : verifyEmail
  const [verifyState, submitVerify, verifyPending] = useActionState(verifyAction, undefined)
  const [sendState, sendAction, sendPending] = useActionState(sendEmailVerification, undefined)

  const title = isChangeMode ? t('auth.verify_email_change_title') : t('auth.verify_email_title')
  const subtitle = isChangeMode ? t('auth.verify_email_change_subtitle') : t('auth.verify_email_subtitle')

  if (verifyState?.success) {
    return (
      <div className="space-y-4 text-center">
        <p className="text-sm text-sidebar-primary">{verifyState.success}</p>
        <Link href="/" className="text-sm text-primary hover:underline">
          {t('auth.verify_email_login_link')}
        </Link>
      </div>
    )
  }

  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">{title}</h2>
        <p className="text-sm text-muted-foreground mt-1">{subtitle}</p>
      </div>

      <div className="space-y-6">
        <form action={submitVerify} className="space-y-4">
          {verifyState?.message && (
            <p role="alert" className="text-sm text-destructive text-center">{verifyState.message}</p>
          )}

          <div className="space-y-1.5">
            <Label htmlFor="email">{t('auth.verify_email_label')}</Label>
            <Input
              id="email"
              name="email"
              type="email"
              autoComplete="email"
              defaultValue={verifyState?.values?.email ?? defaultEmail}
              placeholder={t('auth.verify_email_placeholder')}
              aria-invalid={!!verifyState?.errors?.email}
              aria-describedby={verifyState?.errors?.email ? 'err-email' : undefined}
              required
            />
            {verifyState?.errors?.email && (
              <p id="err-email" role="alert" className="text-xs text-destructive">{verifyState.errors.email[0]}</p>
            )}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="verification_code">{t('auth.verify_code_label')}</Label>
            <Input
              id="verification_code"
              name="verification_code"
              type="text"
              inputMode="numeric"
              maxLength={6}
              placeholder={t('auth.verify_code_placeholder')}
              aria-invalid={!!verifyState?.errors?.verification_code}
              aria-describedby={verifyState?.errors?.verification_code ? 'err-verification_code' : undefined}
              required
            />
            {verifyState?.errors?.verification_code && (
              <p id="err-verification_code" role="alert" className="text-xs text-destructive">{verifyState.errors.verification_code[0]}</p>
            )}
          </div>

          <Button type="submit" className="w-full" disabled={verifyPending}>
            {verifyPending ? t('auth.verify_email_submitting') : t('auth.verify_email_submit')}
          </Button>
        </form>

        <div className="border-t pt-4">
          <p className="text-center text-sm text-muted-foreground mb-3">
            {t('auth.verify_email_no_code')}
          </p>
          {sendState?.success && (
            <p role="status" className="text-xs text-sidebar-primary text-center mb-2">
              {sendState.success}
            </p>
          )}
          <form action={sendAction}>
            <input type="hidden" name="email" value={verifyState?.values?.email ?? defaultEmail ?? ''} />
            <Button type="submit" variant="outline" className="w-full" disabled={sendPending}>
              {sendPending ? t('auth.verify_email_resending') : t('auth.verify_email_resend')}
            </Button>
          </form>
        </div>
      </div>
    </>
  )
}
