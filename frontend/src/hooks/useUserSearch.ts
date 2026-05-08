'use client'
import React from 'react'

export interface UserResult {
  id: number
  username: string
  avatar_url: string
  first_name: string
  last_name: string
}

export function useUserSearch(query: string) {
  const [users, setUsers] = React.useState<UserResult[]>([])
  const [loading, setLoading] = React.useState(false)
  const abortRef = React.useRef<AbortController | null>(null)

  React.useEffect(() => {
    if (!query.trim()) {
      setUsers([])
      setLoading(false)
      return
    }

    if (abortRef.current) abortRef.current.abort()
    const controller = new AbortController()
    abortRef.current = controller

    setLoading(true)
    fetch(`/api/v1/users/search?q=${encodeURIComponent(query.trim())}`, {
      credentials: 'include',
      signal: controller.signal,
    })
      .then((r) => r.json())
      .then((json) => setUsers(json.data?.users ?? []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [query])

  return { users, loading }
}
