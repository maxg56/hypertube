<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

---

# Frontend Conventions

## Stack

| Layer | Technology |
|---|---|
| Framework | Next.js (App Router) — see version note above |
| Language | TypeScript 5 strict mode |
| Styling | Tailwind CSS 4 via PostCSS |
| UI components | shadcn/ui + Radix UI primitives |
| Icons | Lucide React |
| Forms | Native `<form>` with Next.js Server Actions + `useActionState` |
| Validation | Zod (schemas in `src/lib/definitions.ts`) |
| i18n | i18next + react-i18next |
| Auth | httpOnly cookies, managed via Server Actions |

Do not introduce new libraries without explicit approval. Match what is already in use.

---

## Color Palette

The design uses two accent colors on a neutral base. Do not invent additional accent hues.

| Role | Token | Value |
|---|---|---|
| Primary accent | `--sidebar-primary` (dark) | `oklch(0.488 0.243 264.376)` — **blue** |
| Secondary accent / danger | `--destructive` | `oklch(0.577 0.245 27.325)` — **orange** |
| Backgrounds | `--background` | white / dark near-black |
| Text | `--foreground` | near-black / near-white |

Always use CSS variables via Tailwind tokens (`bg-primary`, `text-destructive`, `bg-sidebar-primary`, etc.). Never hard-code hex or rgb values in components.

Dark mode is supported via the `.dark` class. Use `dark:` Tailwind variants when needed.

---

## Internationalization (i18n)

**Every visible string must be translated.** No hard-coded user-facing text is allowed in components.

### Languages

- French (`fr`) — default
- English (`en`) — fallback

### Files

```
src/locales/fr.json   ← source of truth
src/locales/en.json   ← must mirror every key in fr.json
```

### Usage

Only Client Components can call `useTranslation`. If a Server Component needs translated text, pass it as a prop from a Client Component boundary.

```tsx
'use client'
import { useTranslation } from 'react-i18next'

export function MyComponent() {
  const { t } = useTranslation()
  return <p>{t('my_key')}</p>
}
```

When adding a new string:
1. Add the key + French value to `fr.json`
2. Add the same key + English value to `en.json`
3. Use `t('my_key')` in the component — never the raw string

### Key naming

Use `snake_case`. Group related keys with dot notation when the feature grows (e.g. `auth.login_title`, `auth.password_label`).

---

## Component Architecture

### Prefer components over inline markup

Extract any repeated or self-contained UI block into a component. Inline JSX should stay minimal and structural.

### File locations

```
src/components/ui/        ← shadcn/ui base components (Button, Input, Card…)
src/components/auth/      ← feature components for auth flows
src/components/page/      ← layout-level components (Header, Footer…)
src/app/(auth)/           ← route group for auth pages
src/app/(main)/           ← route group for authenticated app
src/app/actions/          ← Server Actions
src/lib/                  ← utilities, schemas, i18n config
src/locales/              ← translation files
```

### Server vs. Client components

- Default to **Server Components**. Add `'use client'` only when you need browser APIs, event handlers, or hooks.
- Server Actions live in `src/app/actions/` and are marked `'use server'`.

### shadcn/ui conventions

- Use the `cn()` utility from `src/lib/utils.ts` to merge Tailwind classes.
- Use CVA (`class-variance-authority`) for variant-based styling.
- Every custom UI component exposes a `className` prop and forwards it via `cn()`.
- Use `data-slot` attributes to enable consistent selector-based styling.

---

## Forms & Validation

- Forms use the native `<form>` element with `action={serverAction}`.
- State is managed with `useActionState` (Next.js built-in).
- Schemas are defined in `src/lib/definitions.ts` using Zod.
- Validation errors are field-level and returned from the Server Action as `ActionState`.
- Display errors directly below the relevant input — never via alert or toast for field errors.

---

## Code Style

- TypeScript strict mode is on. No `any`, no `as unknown as X` casts without justification.
- Functional components only. No class components.
- No default exports for components — use named exports.
- Path alias: `@/*` maps to `src/*`. Always use it instead of relative paths that go up more than one level.
- Do not add comments unless the **why** is non-obvious from the code itself.
