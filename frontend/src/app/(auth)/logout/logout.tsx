'use client'

import { useTranslation } from 'react-i18next'
import { logout } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'

export default function LogoutButton() {
  const { t } = useTranslation()

  return (
    <form action={logout}>
      <Button type="submit" variant="destructive">
        {t('auth.logout')}
      </Button>
    </form>
  )
}
