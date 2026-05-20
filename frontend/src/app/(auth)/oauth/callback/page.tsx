import { redirect } from 'next/navigation'
import { setTokens } from '@/lib/session'

export default async function OAuthCallbackPage({
  searchParams,
}: Readonly<{
  searchParams: Promise<{
    access_token?: string
    refresh_token?: string
    expires_in?: string
    callbackUrl?: string
    error?: string
  }>
}>) {
  const params = await searchParams

  if (params.error) {
    redirect('/login?error=' + encodeURIComponent(params.error))
  }

  if (!params.access_token || !params.refresh_token) {
    redirect('/login?error=oauth_failed')
  }

  await setTokens(
    params.access_token,
    params.refresh_token,
    params.expires_in ? parseInt(params.expires_in, 10) : 900,
  )

  const callbackUrl = params.callbackUrl?.startsWith('/') ? params.callbackUrl : '/'
  redirect(callbackUrl)
}
