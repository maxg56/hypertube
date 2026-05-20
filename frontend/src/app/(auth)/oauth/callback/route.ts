import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'

export async function GET(request: NextRequest) {
  const sp = request.nextUrl.searchParams
  const accessToken = sp.get('access_token')
  const refreshToken = sp.get('refresh_token')
  const expiresIn = parseInt(sp.get('expires_in') ?? '900', 10)
  const callbackUrl = sp.get('callbackUrl') ?? '/'
  const error = sp.get('error')

  const base = request.nextUrl.origin

  if (error || !accessToken || !refreshToken) {
    const dest = new URL('/login', base)
    dest.searchParams.set('error', error ?? 'oauth_failed')
    return NextResponse.redirect(dest)
  }

  const store = await cookies()
  const secure = process.env.NODE_ENV === 'production'

  store.set('access_token', accessToken, {
    httpOnly: true,
    secure,
    sameSite: 'lax',
    path: '/',
    maxAge: expiresIn,
  })

  store.set('refresh_token', refreshToken, {
    httpOnly: true,
    secure,
    sameSite: 'lax',
    path: '/',
    maxAge: 7 * 24 * 60 * 60,
  })

  const safeCallback = callbackUrl.startsWith('/') ? callbackUrl : '/'
  return NextResponse.redirect(new URL(safeCallback, base))
}
