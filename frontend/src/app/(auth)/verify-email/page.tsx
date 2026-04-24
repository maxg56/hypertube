import type { Metadata } from 'next'
import VerifyEmailForm from '@/components/auth/verify-email-form'

export const metadata: Metadata = {
  title: 'Vérification email — Hypertube',
}

export default async function VerifyEmailPage({
  searchParams,
}: {
  searchParams: Promise<{ email?: string }>
}) {
  const { email } = await searchParams

  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">Vérifier votre email</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Entrez le code à 6 chiffres envoyé à votre adresse email.
        </p>
      </div>
      <VerifyEmailForm defaultEmail={email} />
    </>
  )
}
