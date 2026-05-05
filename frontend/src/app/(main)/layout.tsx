import type { ReactNode } from 'react'
import LogoutButton from '@/app/(auth)/logout/logout'

export default function MainLayout({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b bg-background/95 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between">
          <span className="font-bold text-lg tracking-tight">Hypertube</span>
          <LogoutButton />
        </div>
      </header>
      <main className="flex-1 max-w-7xl mx-auto w-full px-4 py-8">
        {children}
      </main>
    </div>
  )
}
