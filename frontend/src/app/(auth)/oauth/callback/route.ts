import { NextRequest, NextResponse } from 'next/server'

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

  const safeCallback = callbackUrl.startsWith('/') ? callbackUrl : '/'
  const response = NextResponse.redirect(new URL(safeCallback, base))
  const secure = process.env.NODE_ENV === 'production'

  response.cookies.set('access_token', accessToken, {
    httpOnly: true,
    secure,
    sameSite: 'lax',
    path: '/',
    maxAge: expiresIn,
  })

  response.cookies.set('refresh_token', refreshToken, {
    httpOnly: true,
    secure,
    sameSite: 'lax',
    path: '/',
    maxAge: 7 * 24 * 60 * 60,
  })

  return response
}
