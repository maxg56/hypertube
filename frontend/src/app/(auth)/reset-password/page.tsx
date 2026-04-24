import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import ResetPasswordForm from '@/components/auth/reset-password-form'

export const metadata: Metadata = {
  title: 'Réinitialisation du mot de passe — Hypertube',
}

export default async function ResetPasswordPage({
  searchParams,
}: {
  searchParams: Promise<{ token?: string }>
}) {
  const { token } = await searchParams

  if (!token) {
    notFound()
  }

  return (
    <>
      <div className="mb-6">
        <h2 className="text-xl font-semibold">Nouveau mot de passe</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Choisissez un nouveau mot de passe sécurisé.
        </p>
      </div>
      <ResetPasswordForm token={token} />
    </>
  )
}
