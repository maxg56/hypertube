'use client'
import type { ReactNode } from 'react'
import { logout } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import LanguageSwitcher from '@/components/page/LanguageSwitcher'
import Link from 'next/link'
import LogoutButton from '@/app/(auth)/logout/logout'
import { usePathname } from 'next/navigation'
import { useEffect, useState } from 'react'

export default function MainLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname()
  const [isProfilePage, setIsProfilePage] = useState(false)

  // Utiliser useEffect pour éviter les problèmes d'hydratation
  useEffect(() => {
    setIsProfilePage(pathname.includes('/profile'))
  }, [pathname])

  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-blue-400 text-white p-4">
            <div className="flex items-center justify-between">
                    {isProfilePage ? (
                        <Link href="/">
                            <Button className="ml-8 bg-white text-blue-400 hover:bg-gray-100">
                                Homepage
                            </Button>
                        </Link>
                    ) : (
                        <Link href="/profile">
                            <Avatar className="size-20 ml-8 bg-gray-300 cursor-pointer hover:opacity-80 transition-opacity">
                                <AvatarImage
                                    src={`https://robohash.org/1.png?set=set1`}
                                    alt="Avatar"
                                />
                                <AvatarFallback>HT</AvatarFallback>
                            </Avatar>
                        </Link>
                    )}
                <h1 className="text-2xl text-gray-800 font-bold text-center flex-1">Hypertube</h1>
                <div className="flex items-center gap-4">
                    <LanguageSwitcher />
                    <LogoutButton />
                </div>
            </div>            
        </header>
      <main className="w-full bg-gradient-to-t from-orange-400 to-blue-400 p-6 shadow-lg min-h-screen">
        {children}
      </main>
          <footer className="bg-orange-400 text-white p-4 mt-auto">
            <p className="text-center">&copy; 2024 Hypertube. Tous droits réservés.</p>
          </footer>
    </div>
  )
}
