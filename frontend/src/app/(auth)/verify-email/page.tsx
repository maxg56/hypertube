import type { Metadata } from 'next'
import VerifyEmailForm from '@/components/auth/verify-email-form'

export const metadata: Metadata = {
  title: 'Vérification email — Hypertube',
}

export default async function VerifyEmailPage({
  searchParams,
}: Readonly<{
  searchParams: Promise<{ email?: string; mode?: string }>
}>) {
  const { email, mode } = await searchParams

  return <VerifyEmailForm defaultEmail={email} mode={mode} />
}
