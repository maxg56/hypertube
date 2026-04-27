'use client'

import { createContext, useContext, useState, useCallback } from 'react'

interface AuthContextType {
  onLoginSuccess: (callback: () => void) => void
  triggerLoginSuccess: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [loginSuccessCallback, setLoginSuccessCallback] = useState<(() => void) | null>(null)

  const onLoginSuccess = useCallback((callback: () => void) => {
    setLoginSuccessCallback(() => callback)
  }, [])

  const triggerLoginSuccess = useCallback(() => {
    if (loginSuccessCallback) {
      loginSuccessCallback()
    }
  }, [loginSuccessCallback])

  return (
    <AuthContext.Provider value={{ onLoginSuccess, triggerLoginSuccess }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuthContext() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuthContext doit être utilisé dans AuthProvider')
  return context
}
