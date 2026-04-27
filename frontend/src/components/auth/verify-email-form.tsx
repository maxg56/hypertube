'use client'

import { useActionState } from 'react'
import Link from 'next/link'
import { verifyEmail, sendEmailVerification } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function VerifyEmailForm({ defaultEmail }: { defaultEmail?: string }) {
  const [verifyState, verifyAction, verifyPending] = useActionState(
    verifyEmail,
    undefined,
  )
  const [sendState, sendAction, sendPending] = useActionState(
    sendEmailVerification,
    undefined,
  )

  if (verifyState?.success) {
    return (
      <div className="space-y-4 text-center">
        <p className="text-sm text-green-600 dark:text-green-400">
          {verifyState.success}
        </p>
        <Link href="/login" className="text-sm text-primary hover:underline">
          Se connecter
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <form action={verifyAction} className="space-y-4">
        {verifyState?.message && (
          <p className="text-sm text-destructive text-center">
            {verifyState.message}
          </p>
        )}

        <div className="space-y-1.5">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            defaultValue={defaultEmail}
            placeholder="john@example.com"
            required
          />
          {verifyState?.errors?.email && (
            <p className="text-xs text-destructive">
              {verifyState.errors.email[0]}
            </p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="verification_code">Code de vérification</Label>
          <Input
            id="verification_code"
            name="verification_code"
            type="text"
            inputMode="numeric"
            maxLength={6}
            placeholder="123456"
            required
          />
          {verifyState?.errors?.verification_code && (
            <p className="text-xs text-destructive">
              {verifyState.errors.verification_code[0]}
            </p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={verifyPending}>
          {verifyPending ? 'Vérification…' : 'Vérifier'}
        </Button>
      </form>

      <div className="border-t pt-4">
        <p className="text-center text-sm text-muted-foreground mb-3">
          Pas reçu le code ?
        </p>
        {sendState?.success && (
          <p className="text-xs text-green-600 dark:text-green-400 text-center mb-2">
            {sendState.success}
          </p>
        )}
        <form action={sendAction}>
          <input type="hidden" name="email" value={defaultEmail ?? ''} />
          <Button
            type="submit"
            variant="outline"
            className="w-full"
            disabled={sendPending}
          >
            {sendPending ? 'Envoi…' : 'Renvoyer le code'}
          </Button>
        </form>
      </div>
    </div>
  )
}
