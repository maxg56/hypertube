import { NextRequest, NextResponse } from 'next/server'

const PUBLIC_ROUTES = [
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/verify-email',
]

// Routes that are public but must never redirect authenticated users away
// (e.g. OAuth callback needs to run regardless of auth state)
const ALWAYS_PUBLIC_ROUTES = ['/oauth/callback']

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl
  const isAlwaysPublic = ALWAYS_PUBLIC_ROUTES.some((r) => pathname.startsWith(r))
  if (isAlwaysPublic) return NextResponse.next()

  const isPublic = PUBLIC_ROUTES.some((r) => pathname.startsWith(r))
  const accessToken = req.cookies.get('access_token')?.value

  if (!isPublic && !accessToken) {
    const loginUrl = new URL('/login', req.nextUrl)
    loginUrl.searchParams.set('callbackUrl', pathname)
    return NextResponse.redirect(loginUrl)
  }

  if (isPublic && accessToken) {
    return NextResponse.redirect(new URL('/', req.nextUrl))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next|.*\\.ico$|.*\\.png$|.*\\.svg$).*)'],
}
