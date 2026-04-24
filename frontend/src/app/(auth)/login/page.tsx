import type { Metadata } from 'next'
import LoginForm from '@/components/auth/login-form'

export const metadata: Metadata = {
  title: 'Connexion — Hypertube',
}

export default function LoginPage() {
  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">Connexion</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Bienvenue ! Entrez vos identifiants pour continuer.
        </p>
      </div>
      <LoginForm />
    </>
  )
}
