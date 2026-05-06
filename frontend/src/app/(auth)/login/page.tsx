import type { Metadata } from 'next'
import LoginForm from '@/components/auth/login-form'

export const metadata: Metadata = {
  title: 'Connexion — Hypertube',
}

export default async function LoginPage({
  searchParams,
}: Readonly<{
  searchParams: Promise<{ callbackUrl?: string }>
}>) {
  const { callbackUrl } = await searchParams

  return <LoginForm callbackUrl={callbackUrl} />
}
