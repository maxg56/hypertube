import type { Metadata } from 'next'
import ForgotPasswordForm from '@/components/auth/forgot-password-form'

export const metadata: Metadata = {
  title: 'Mot de passe oublié — Hypertube',
}

export default function ForgotPasswordPage() {
  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">Mot de passe oublié</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Entrez votre email pour recevoir un lien de réinitialisation.
        </p>
      </div>
      <ForgotPasswordForm />
    </>
  )
}
