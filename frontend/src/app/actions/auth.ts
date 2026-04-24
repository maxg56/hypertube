'use server'

import { redirect } from 'next/navigation'
import {
  LoginSchema,
  RegisterSchema,
  ForgotPasswordSchema,
  ResetPasswordSchema,
  VerifyEmailSchema,
  SendVerificationSchema,
  type ActionState,
} from '@/lib/definitions'
import { setTokens, clearTokens, getAccessToken, getRefreshToken } from '@/lib/session'

const API = process.env.API_URL ?? 'http://localhost:8080'

async function apiFetch(path: string, options: RequestInit = {}) {
  const { headers = {}, ...rest } = options
  return fetch(`${API}${path}`, {
    ...rest,
    headers: { 'Content-Type': 'application/json', ...(headers as Record<string, string>) },
  })
}

export async function login(_state: ActionState, formData: FormData): Promise<ActionState> {
  const validated = LoginSchema.safeParse({
    login: formData.get('login'),
    password: formData.get('password'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten().fieldErrors }
  }

  const res = await apiFetch('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify(validated.data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.message ?? 'Identifiants invalides' }
  }

  const { data } = await res.json()
  await setTokens(data.access_token, data.refresh_token, data.expires_in)
  redirect('/')
}

export async function register(_state: ActionState, formData: FormData): Promise<ActionState> {
  const validated = RegisterSchema.safeParse({
    username: formData.get('username'),
    email: formData.get('email'),
    password: formData.get('password'),
    first_name: formData.get('first_name'),
    last_name: formData.get('last_name'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten().fieldErrors }
  }

  const res = await apiFetch('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify(validated.data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.message ?? "L'inscription a échoué" }
  }

  const { data } = await res.json()
  await setTokens(data.access_token, data.refresh_token, data.expires_in)
  redirect('/')
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
  redirect('/login')
}

export async function forgotPassword(
  _state: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const validated = ForgotPasswordSchema.safeParse({
    email: formData.get('email'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten().fieldErrors }
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
    return { errors: validated.error.flatten().fieldErrors }
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
    return { message: body.message ?? 'Lien invalide ou expiré' }
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
    return { errors: validated.error.flatten().fieldErrors }
  }

  const res = await apiFetch('/api/v1/auth/send-email-verification', {
    method: 'POST',
    body: JSON.stringify(validated.data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.message ?? "Erreur lors de l'envoi" }
  }

  return { success: 'Code envoyé. Vérifiez votre boîte mail.' }
}

export async function verifyEmail(
  _state: ActionState,
  formData: FormData,
): Promise<ActionState> {
  const validated = VerifyEmailSchema.safeParse({
    email: formData.get('email'),
    verification_code: formData.get('verification_code'),
  })

  if (!validated.success) {
    return { errors: validated.error.flatten().fieldErrors }
  }

  const res = await apiFetch('/api/v1/auth/verify-email', {
    method: 'POST',
    body: JSON.stringify(validated.data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    return { message: body.message ?? 'Code invalide ou expiré' }
  }

  return { success: 'Email vérifié avec succès !' }
}
