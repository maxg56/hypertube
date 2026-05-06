import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import ResetPasswordForm from '@/components/auth/reset-password-form'

export const metadata: Metadata = {
  title: 'Réinitialisation du mot de passe — Hypertube',
}

export default async function ResetPasswordPage({
  searchParams,
}: Readonly<{
  searchParams: Promise<{ token?: string }>
}>) {
  const { token } = await searchParams

  if (!token) {
    notFound()
  }

  return <ResetPasswordForm token={token} />
}
