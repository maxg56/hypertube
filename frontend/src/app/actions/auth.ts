'use server'

import { redirect } from 'next/navigation'
import {
  LoginSchema,
  RegisterSchema,
  ForgotPasswordSchema,
  ResetPasswordSchema,
  VerifyEmailSchema,
  SendVerificationSchema,
  RequestEmailChangeSchema,
  ConfirmEmailChangeSchema,
  type ActionState,
} from '@/lib/definitions'
import { setTokens, setEmailVerified, clearTokens, getAccessToken, getRefreshToken } from '@/lib/session'

const API = process.env.API_URL ?? 'https://localhost:8443'

async function apiFetch(path: string, options: RequestInit = {}) {
  const { headers = {}, ...rest } = options
  return fetch(`${API}${path}`, {
    ...rest,
    headers: { 'Content-Type': 'application/json', ...(headers as Record<string, string>) },
  })
}

export async function login(_state: ActionState, formData: FormData): Promise<ActionState> {
  const loginValue = formData.get('login') as string
  const values = { login: loginValue }

  const validated = LoginSchema.safeParse({
    login: loginValue,
    password: formData.get('password'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten(i => i.message).fieldErrors, values }
  }

  const res = await fetch(`${API}/api/v1/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'password',
      username: validated.data.login,
      password: validated.data.password,
    }).toString(),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.error_description ?? body.error ?? 'Identifiants invalides', values }
  }

  const data = await res.json()
  await setTokens(data.access_token, data.refresh_token, data.expires_in)
  await setEmailVerified(data.email_verified === true)

  if (!data.email_verified) {
    redirect(`/verify-email?email=${encodeURIComponent(data.email ?? '')}`)
  }

  const callbackUrl = formData.get('callbackUrl') as string | null
  redirect(callbackUrl?.startsWith('/') ? callbackUrl : '/')
}

export async function register(_state: ActionState, formData: FormData): Promise<ActionState> {
  const raw = {
    username: formData.get('username') as string,
    email: formData.get('email') as string,
    password: formData.get('password') as string,
    confirm_password: formData.get('confirm_password') as string,
    first_name: formData.get('first_name') as string,
    last_name: formData.get('last_name') as string,
  }

  const values = {
    username: raw.username,
    email: raw.email,
    first_name: raw.first_name,
    last_name: raw.last_name,
  }

  const validated = RegisterSchema.safeParse(raw)

  if (!validated.success) {
    return { errors: validated.error.flatten(i => i.message).fieldErrors, values }
  }

  const { confirm_password: _, ...payload } = validated.data
  const res = await apiFetch('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.error ?? "L'inscription a échoué", values }
  }

  const { data } = await res.json()
  await setTokens(data.access_token, data.refresh_token, data.expires_in)
  await setEmailVerified(false)
  redirect(`/verify-email?email=${encodeURIComponent(data.user?.email ?? '')}`)
}

export async function refreshToken(): Promise<boolean> {
  const token = await getRefreshToken()
  if (!token) return false

  const res = await fetch(`${API}/api/v1/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'refresh_token',
      refresh_token: token,
    }).toString(),
  })

  if (!res.ok) {
    await clearTokens()
    return false
  }

  const data = await res.json()
  await setTokens(data.access_token, data.refresh_token, data.expires_in)
  return true
}

export async function logout() {
  const accessToken = await getAccessToken()
  const refreshToken = await getRefreshToken()

  if (accessToken) {
    await apiFetch('/api/v1/auth/logout', {
      method: 'POST',
      headers: { Authorization: `Bearer ${accessToken}` },
      body: JSON.stringify({ refresh_token: refreshToken }),
    }).catch(() => {})
  }

  await clearTokens()
}

export async function forgotPassword(
  _state: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const email = formData.get('email') as string
  const values = { email }

  const validated = ForgotPasswordSchema.safeParse({ email })

  if (!validated.success) {
    return { errors: validated.error.flatten(i => i.message).fieldErrors, values }
  }

  await apiFetch('/api/v1/auth/forgot-password', {
    method: 'POST',
    body: JSON.stringify(validated.data),
  }).catch(() => {})

  return {
    success:
      'Si cet email existe, un lien de réinitialisation vous a été envoyé.',
  }
}

export async function resetPassword(
  _state: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const validated = ResetPasswordSchema.safeParse({
    token: formData.get('token'),
    new_password: formData.get('new_password'),
    confirm_password: formData.get('confirm_password'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten(i => i.message).fieldErrors }
  }

  const res = await apiFetch('/api/v1/auth/reset-password', {
    method: 'POST',
    body: JSON.stringify({
      token: validated.data.token,
      new_password: validated.data.new_password,
    }),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.error ??'Lien invalide ou expiré' }
  }

  redirect('/login')
}

export async function sendEmailVerification(
  _state: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const validated = SendVerificationSchema.safeParse({
    email: formData.get('email'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten(i => i.message).fieldErrors }
  }

  const res = await apiFetch('/api/v1/auth/send-email-verification', {
    method: 'POST',
    body: JSON.stringify(validated.data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.error ??"Erreur lors de l'envoi" }
  }

  return { success: 'Code envoyé. Vérifiez votre boîte mail.' }
}

export async function verifyEmail(
  _state: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const email = formData.get('email') as string
  const values = { email }

  const validated = VerifyEmailSchema.safeParse({
    email,
    verification_code: formData.get('verification_code'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten(i => i.message).fieldErrors, values }
  }

  const res = await apiFetch('/api/v1/auth/verify-email', {
    method: 'POST',
    body: JSON.stringify(validated.data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.error ?? 'Code invalide ou expiré', values }
  }

  await setEmailVerified(true)
  redirect('/')
}

export async function requestEmailChange(
  _state: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const validated = RequestEmailChangeSchema.safeParse({
    new_email: formData.get('new_email'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten(i => i.message).fieldErrors }
  }

  const accessToken = await getAccessToken()
  const res = await apiFetch('/api/v1/auth/request-email-change', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(validated.data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.error ?? "Impossible d'envoyer le code de vérification" }
  }

  redirect(`/verify-email?email=${encodeURIComponent(validated.data.new_email)}&mode=change`)
}

export async function confirmEmailChange(
  _state: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const validated = ConfirmEmailChangeSchema.safeParse({
    new_email: formData.get('email'),
    verification_code: formData.get('verification_code'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten(i => i.message).fieldErrors }
  }

  const accessToken = await getAccessToken()
  const res = await apiFetch('/api/v1/auth/confirm-email-change', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(validated.data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.error ?? 'Code invalide ou expiré' }
  }

  await setEmailVerified(true)
  redirect('/profile')
}
