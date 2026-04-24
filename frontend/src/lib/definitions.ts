import { z } from 'zod'

export const LoginSchema = z.object({
  login: z.string().min(1, { message: 'Ce champ est requis' }),
  password: z.string().min(1, { message: 'Ce champ est requis' }),
})

export const RegisterSchema = z.object({
  username: z
    .string()
    .min(3, { message: 'Au moins 3 caractères' })
    .max(50, { message: '50 caractères maximum' }),
  email: z.string().email({ message: 'Email invalide' }),
  password: z
    .string()
    .min(8, { message: 'Au moins 8 caractères' })
    .regex(/[a-zA-Z]/, { message: 'Au moins une lettre' })
    .regex(/[0-9]/, { message: 'Au moins un chiffre' }),
  first_name: z.string().min(1, { message: 'Ce champ est requis' }),
  last_name: z.string().min(1, { message: 'Ce champ est requis' }),
})

export const ForgotPasswordSchema = z.object({
  email: z.string().email({ message: 'Email invalide' }),
})

export const ResetPasswordSchema = z
  .object({
    token: z.string().min(1),
    new_password: z.string().min(8, { message: 'Au moins 8 caractères' }),
    confirm_password: z.string(),
  })
  .refine((d) => d.new_password === d.confirm_password, {
    message: 'Les mots de passe ne correspondent pas',
    path: ['confirm_password'],
  })

export const VerifyEmailSchema = z.object({
  email: z.string().email({ message: 'Email invalide' }),
  verification_code: z
    .string()
    .length(6, { message: 'Code à 6 chiffres requis' }),
})

export const SendVerificationSchema = z.object({
  email: z.string().email({ message: 'Email invalide' }),
})

export type ActionState =
  | {
      errors?: Record<string, string[]>
      message?: string
      success?: string
    }
  | undefined
