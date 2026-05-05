import type { ReactNode } from 'react'
import { logout } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import LanguageSwitcher from '@/components/page/LanguageSwitcher'
import Link from 'next/link'
import LogoutButton from '@/app/(auth)/logout/logout'

export default function MainLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-blue-400 text-white p-4">
            <div className="flex items-center justify-between">
                    <Link href="/profile">
                        <Avatar className="size-20 ml-8 bg-gray-300 cursor-pointer hover:opacity-80 transition-opacity">
                            <AvatarImage
                                src={`https://robohash.org/1.png?set=set1`}
                                alt="Avatar"
                            />
                            <AvatarFallback>HT</AvatarFallback>
                        </Avatar>
                    </Link>
                <h1 className="text-2xl text-gray-800 font-bold text-center flex-1">Hypertube</h1>
                <div className="flex items-center gap-4">
                    <LanguageSwitcher />
                    <LogoutButton />
                </div>
            </div>            
        </header>
      <main className="flex-1 max-w-7xl mx-auto w-full px-4 py-8">
        {children}
      </main>
    </div>
  )
}
