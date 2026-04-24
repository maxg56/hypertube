'use client'

import { useActionState } from 'react'
import Link from 'next/link'
import { register } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function RegisterForm() {
  const [state, action, pending] = useActionState(register, undefined)

  return (
    <form action={action} className="space-y-4">
      {state?.message && (
        <p className="text-sm text-destructive text-center">{state.message}</p>
      )}

      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="first_name">Prénom</Label>
          <Input
            id="first_name"
            name="first_name"
            type="text"
            autoComplete="given-name"
            placeholder="Jean"
            required
          />
          {state?.errors?.first_name && (
            <p className="text-xs text-destructive">
              {state.errors.first_name[0]}
            </p>
          )}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="last_name">Nom</Label>
          <Input
            id="last_name"
            name="last_name"
            type="text"
            autoComplete="family-name"
            placeholder="Dupont"
            required
          />
          {state?.errors?.last_name && (
            <p className="text-xs text-destructive">
              {state.errors.last_name[0]}
            </p>
          )}
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="username">Nom d'utilisateur</Label>
        <Input
          id="username"
          name="username"
          type="text"
          autoComplete="username"
          placeholder="johndoe"
          required
        />
        {state?.errors?.username && (
          <p className="text-xs text-destructive">
            {state.errors.username[0]}
          </p>
        )}
      </div>

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

      <div className="space-y-1.5">
        <Label htmlFor="password">Mot de passe</Label>
        <Input
          id="password"
          name="password"
          type="password"
          autoComplete="new-password"
          placeholder="••••••••"
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
        {pending ? 'Inscription…' : "S'inscrire"}
      </Button>

      <p className="text-center text-sm text-muted-foreground">
        Déjà un compte ?{' '}
        <Link href="/login" className="text-primary hover:underline">
          Se connecter
        </Link>
      </p>
    </form>
  )
}
