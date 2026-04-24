import type { ReactNode } from 'react'

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <div className="w-full max-w-md space-y-6">
        <div className="text-center">
          <h1 className="text-3xl font-bold tracking-tight">Hypertube</h1>
          <p className="text-muted-foreground text-sm mt-1">
            Regardez vos films en streaming
          </p>
        </div>
        <div className="bg-card border rounded-xl p-6 shadow-sm">
          {children}
        </div>
      </div>
    </div>
  )
}
