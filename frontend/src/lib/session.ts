import 'server-only'
import { cookies } from 'next/headers'

const ACCESS_TOKEN = 'access_token'
const REFRESH_TOKEN = 'refresh_token'
const EMAIL_VERIFIED = 'email_verified'

export async function setTokens(
  accessToken: string,
  refreshToken: string,
  expiresIn: number,
) {
  const store = await cookies()
  const secure = process.env.NODE_ENV === 'production'

  store.set(ACCESS_TOKEN, accessToken, {
    httpOnly: true,
    secure,
    sameSite: 'lax',
    path: '/',
    maxAge: expiresIn,
  })

  store.set(REFRESH_TOKEN, refreshToken, {
    httpOnly: true,
    secure,
    sameSite: 'lax',
    path: '/',
    maxAge: 7 * 24 * 60 * 60,
  })
}

export async function setEmailVerified(verified: boolean) {
  const store = await cookies()
  const secure = process.env.NODE_ENV === 'production'
  store.set(EMAIL_VERIFIED, String(verified), {
    httpOnly: true,
    secure,
    sameSite: 'lax',
    path: '/',
    maxAge: 7 * 24 * 60 * 60,
  })
}

export async function getAccessToken() {
  const store = await cookies()
  return store.get(ACCESS_TOKEN)?.value
}

export async function getRefreshToken() {
  const store = await cookies()
  return store.get(REFRESH_TOKEN)?.value
}

export async function clearTokens() {
  const store = await cookies()
  store.delete(ACCESS_TOKEN)
  store.delete(REFRESH_TOKEN)
  store.delete(EMAIL_VERIFIED)
}
