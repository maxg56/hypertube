'use client'
import React from 'react'
import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Loader2, ArrowLeft } from 'lucide-react'

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const { t } = useTranslation()
  const pathname = usePathname()
  const router = useRouter()
  const [ready, setReady] = React.useState(false)

  React.useEffect(() => {
    fetch('/api/v1/users/profile', { credentials: 'include' })
      .then((r) => r.json())
      .then(({ data }) => {
        if (data?.profile?.role !== 'admin') {
          router.replace('/')
        } else {
          setReady(true)
        }
      })
      .catch(() => router.replace('/'))
  }, [router])

  if (!ready) {
    return (
      <div className="flex items-center justify-center min-h-[40vh]">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const tabs = [
    { href: '/admin/overview', label: t('admin.tab_overview') },
    { href: '/admin/users', label: t('admin.tab_users') },
    { href: '/admin/films', label: t('admin.tab_films') },
  ]

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="flex items-center gap-4 mb-6">
        <Link href="/" className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors">
          <ArrowLeft className="size-4" />
          Hypertube
        </Link>
        <h1 className="text-3xl font-bold">{t('admin.title')}</h1>
      </div>
      <nav className="flex gap-1 mb-6 border-b">
        {tabs.map((tab) => {
          const active = pathname.startsWith(tab.href)
          return (
            <Link
              key={tab.href}
              href={tab.href}
              className={cn(
                'px-4 py-2 text-sm font-medium rounded-t-md transition-colors -mb-px border-b-2',
                active
                  ? 'border-sidebar-primary text-sidebar-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border',
              )}
            >
              {tab.label}
            </Link>
          )
        })}
      </nav>
      {children}
    </div>
  )
}
