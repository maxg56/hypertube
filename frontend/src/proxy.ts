import { NextRequest, NextResponse } from 'next/server'

const PUBLIC_ROUTES = [
  '/',
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/verify-email',
]

export default function proxy(req: NextRequest) {
  const { pathname } = req.nextUrl
  const isPublic = PUBLIC_ROUTES.some((r) => pathname.startsWith(r))
  const accessToken = req.cookies.get('access_token')?.value

  if (!isPublic && !accessToken) {
    return NextResponse.redirect(new URL('/login', req.nextUrl))
  }

  if (isPublic && accessToken) {
    return NextResponse.redirect(new URL('/', req.nextUrl))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|.*\\.ico$|.*\\.png$|.*\\.svg$).*)'],
}
