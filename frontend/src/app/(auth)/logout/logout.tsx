'use client'

import { useTransition } from 'react'
import { logout } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { LogOut } from 'lucide-react'

export default function LogoutButton() {
  const [isPending, startTransition] = useTransition()

  return (
    <Button
      disabled={isPending}
      onClick={() => startTransition(() => logout())}
      className="bg-orange-400 hover:bg-orange-500"
    >
      <LogOut className="size-8" />
    </Button>
  )
}
