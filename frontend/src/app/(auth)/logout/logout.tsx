'use client'

import { useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { logout } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { LogOut } from 'lucide-react'

export default function LogoutButton() {
  const [isPending, startTransition] = useTransition()
  const router = useRouter()

  return (
    <Button
      disabled={isPending}
      onClick={() => startTransition(async () => {
        await logout()
        router.push('/login')
      })}
      className="bg-orange-400 hover:bg-orange-500"
    >
      <LogOut className="size-8" />
    </Button>
  )
}
