'use client'

import { useActionState } from 'react'
import Link from 'next/link'
import { forgotPassword } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function ForgotPasswordForm() {
  const [state, action, pending] = useActionState(forgotPassword, undefined)

  if (state?.success) {
    return (
      <div className="space-y-4 text-center">
        <p className="text-sm text-green-600 dark:text-green-400">
          {state.success}
        </p>
        <Link href="/login" className="text-sm text-primary hover:underline">
          Retour à la connexion
        </Link>
      </div>
    )
  }

  return (
    <form action={action} className="space-y-4">
      {state?.message && (
        <p className="text-sm text-destructive text-center">{state.message}</p>
      )}

      <div className="space-y-1.5">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          name="email"
          type="email"
          autoComplete="email"
          placeholder="john@example.com"
          required
        />
        {state?.errors?.email && (
          <p className="text-xs text-destructive">{state.errors.email[0]}</p>
        )}
      </div>

      <Button type="submit" className="w-full" disabled={pending}>
        {pending ? 'Envoi…' : 'Envoyer le lien'}
      </Button>

      <p className="text-center text-sm text-muted-foreground">
        <Link href="/login" className="text-primary hover:underline">
          Retour à la connexion
        </Link>
      </p>
    </form>
  )
}
