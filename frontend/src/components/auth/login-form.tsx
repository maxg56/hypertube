'use client'

import { useActionState, useEffect } from 'react'
import Link from 'next/link'
import { login } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuthContext } from '@/context/auth-context'

export default function LoginForm({ onSuccess }: { onSuccess?: () => void }) {
  const [state, action, pending] = useActionState(login, undefined)
  const { triggerLoginSuccess } = useAuthContext()

  useEffect(() => {
    if (state && !state.message && !state.errors) {
      onSuccess?.()
      triggerLoginSuccess()
    }
  }, [state, onSuccess, triggerLoginSuccess])

  return (
    <form action={action} className="space-y-4">
      {state?.message && (
        <p className="text-sm text-destructive text-center">{state.message}</p>
      )}

      <div className="space-y-1.5">
        <Label htmlFor="login">Nom d'utilisateur ou email</Label>
        <Input
          id="login"
          name="login"
          type="text"
          autoComplete="username"
          placeholder="johndoe ou john@example.com"
          required
        />
        {state?.errors?.login && (
          <p className="text-xs text-destructive">{state.errors.login[0]}</p>
        )}
      </div>

      <div className="space-y-1.5">
        <div className="flex items-center justify-between">
          <Label htmlFor="password">Mot de passe</Label>
          <Link
            href="/forgot-password"
            className="text-xs text-muted-foreground hover:text-primary transition-colors"
          >
            Mot de passe oublié ?
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
        {pending ? 'Connexion…' : 'Se connecter'}
      </Button>

      <p className="text-center text-sm text-muted-foreground">
        Pas encore de compte ?{' '}
        <Link href="/register" className="text-primary hover:underline">
          S'inscrire
        </Link>
      </p>
    </form>
  )
}
