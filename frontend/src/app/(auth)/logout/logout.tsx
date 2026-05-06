'use client'

import { useTranslation } from 'react-i18next'
import { logout } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import {LogOut } from 'lucide-react'

export default function LogoutButton() {
  const { t } = useTranslation()

  return (
    <form action={logout}>
      <Button type="submit" className='bg-orange-400 hover:bg-orange-500'>
        <LogOut className="size-8" />
      </Button>
    </form>
  )
}
