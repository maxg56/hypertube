'use client'
import type { ReactNode } from 'react'
import LogoutButton from '@/app/(auth)/logout/logout'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import LanguageSwitcher from '@/components/page/LanguageSwitcher'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useTranslation } from 'react-i18next'
import { ThemeToggle } from '@/components/page/ThemeToggle'

export default function MainLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname()
  const { t } = useTranslation()
  const isProfilePage = pathname.includes('/profile')

  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b bg-background/95 backdrop-blur-sm sticky top-0 z-50">
        <div className="px-8 h-16 flex items-center justify-between gap-4">
          <ThemeToggle />
          {isProfilePage ? (
            <Link href="/">
              <Button variant="outline" size="sm">{t('nav.home')}</Button>
            </Link>
          ) : (
            <Link href="/profile">
              <Avatar className="size-10 cursor-pointer hover:opacity-80 transition-opacity">
                <AvatarImage src="https://robohash.org/1.png?set=set1" alt="Avatar" />
                <AvatarFallback>HT</AvatarFallback>
              </Avatar>
            </Link>
          )}
          <span className="font-bold text-lg tracking-tight flex-1 text-center">Hypertube</span>
          <div className="flex items-center gap-3">
            <LanguageSwitcher />
            <LogoutButton />
          </div>
        </div>
      </header>
      <main className="flex-1">
        {children}
      </main>
      <footer className="border-t bg-background">
        <div className="max-w-7xl mx-auto px-4 py-4 text-center text-sm text-muted-foreground">
          &copy; {new Date().getFullYear()} Hypertube
        </div>
      </footer>
    </div>
  )
}
