import { NextRequest, NextResponse } from 'next/server'

const PUBLIC_ROUTES = [
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
]

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl
  const isPublic = PUBLIC_ROUTES.some((r) => pathname.startsWith(r))
  const accessToken = req.cookies.get('access_token')?.value
  const emailVerified = req.cookies.get('email_verified')?.value
  const isVerifyEmail = pathname.startsWith('/verify-email')
  const isEmailChangeMode = req.nextUrl.searchParams.get('mode') === 'change'

  // Email change confirmation: logged-in users always allowed through
  if (isVerifyEmail && isEmailChangeMode) {
    if (!accessToken) {
      return NextResponse.redirect(new URL('/login', req.nextUrl))
    }
    return NextResponse.next()
  }

  if (!accessToken) {
    if (!isPublic && !isVerifyEmail) {
      const loginUrl = new URL('/login', req.nextUrl)
      loginUrl.searchParams.set('callbackUrl', pathname)
      return NextResponse.redirect(loginUrl)
    }
    return NextResponse.next()
  }

  // Logged in but email not verified: gate to /verify-email only
  if (emailVerified === 'false') {
    if (isVerifyEmail) return NextResponse.next()
    return NextResponse.redirect(new URL('/verify-email', req.nextUrl))
  }

  // Logged in and verified: redirect away from public/verify-email pages
  if (isPublic || isVerifyEmail) {
    return NextResponse.redirect(new URL('/', req.nextUrl))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next|.*\\.ico$|.*\\.png$|.*\\.svg$).*)'],
}
