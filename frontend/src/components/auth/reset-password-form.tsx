'use client'

import { useActionState } from 'react'
import Link from 'next/link'
import { resetPassword } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function ResetPasswordForm({ token }: { token: string }) {
  const [state, action, pending] = useActionState(resetPassword, undefined)

  return (
    <form action={action} className="space-y-4">
      <input type="hidden" name="token" value={token} />

      {state?.message && (
        <p className="text-sm text-destructive text-center">{state.message}</p>
      )}

      <div className="space-y-1.5">
        <Label htmlFor="new_password">Nouveau mot de passe</Label>
        <Input
          id="new_password"
          name="new_password"
          type="password"
          autoComplete="new-password"
          placeholder="••••••••"
          required
        />
        {state?.errors?.new_password && (
          <p className="text-xs text-destructive">
            {state.errors.new_password[0]}
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="confirm_password">Confirmer le mot de passe</Label>
        <Input
          id="confirm_password"
          name="confirm_password"
          type="password"
          autoComplete="new-password"
          placeholder="••••••••"
          required
        />
        {state?.errors?.confirm_password && (
          <p className="text-xs text-destructive">
            {state.errors.confirm_password[0]}
          </p>
        )}
      </div>

      <Button type="submit" className="w-full" disabled={pending}>
        {pending ? 'Réinitialisation…' : 'Réinitialiser'}
      </Button>

      <p className="text-center text-sm text-muted-foreground">
        <Link href="/login" className="text-primary hover:underline">
          Retour à la connexion
        </Link>
      </p>
    </form>
  )
}
