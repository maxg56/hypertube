import type { Metadata } from 'next'
import RegisterForm from '@/components/auth/register-form'

export const metadata: Metadata = {
  title: 'Inscription — Hypertube',
}

export default function RegisterPage() {
  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">Créer un compte</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Rejoignez Hypertube pour commencer à regarder.
        </p>
      </div>
      <RegisterForm />
    </>
  )
}
