import { useRouter } from 'expo-router'
import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'

import {
  api,
  clearTokens,
  getAccessToken,
  LOGIN_PATH,
  onAuthFailure,
  refreshTokens,
  setServerUrl,
  setTokens,
  type LoginResponse,
  type User,
} from './api'

interface LoginInput {
  serverUrl: string
  username: string
  password: string
}

interface AuthContextValue {
  user: User | null
  isLoading: boolean
  login: (input: LoginInput) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const router = useRouter()
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    return onAuthFailure(() => {
      setUser(null)
      router.replace('/login')
    })
  }, [router])

  // On launch: a stored token is verified by refreshing it. The refresh response is
  // expected to carry the user alongside the tokens.
  useEffect(() => {
    let cancelled = false
    async function restore() {
      try {
        if (!(await getAccessToken())) return
        const { user } = await refreshTokens()
        if (!cancelled && user) setUser(user)
      } catch {
        await clearTokens()
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }
    void restore()
    return () => {
      cancelled = true
    }
  }, [])

  const login = useCallback(async ({ serverUrl, username, password }: LoginInput) => {
    if (serverUrl.trim()) await setServerUrl(serverUrl)
    const res = await api.post<LoginResponse>(LOGIN_PATH, { username, password }, { skipAuthRefresh: true })
    await setTokens(res.data.access_token, res.data.refresh_token)
    setUser(res.data.user)
  }, [])

  const logout = useCallback(async () => {
    await clearTokens()
    setUser(null)
  }, [])

  const value = useMemo(() => ({ user, isLoading, login, logout }), [user, isLoading, login, logout])
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}
