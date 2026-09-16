import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import {
  api,
  clearUser,
  loadUser,
  saveUser,
  type UpdateProfileInput,
} from '../lib/api'
import type { User } from '../lib/types'

type AuthContextValue = {
  user: User | null
  loading: boolean
  error: string
  login: (username: string, password: string) => Promise<void>
  register: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  updateProfile: (patch: UpdateProfileInput) => Promise<void>
  updateAvatar: (file: File) => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false

    async function boot() {
      const cached = loadUser()

      try {
        await api.verify()
      } catch {
        try {
          await api.refresh()
          await api.verify()
        } catch {
          clearUser()
          if (!cancelled) setUser(null)
          if (!cancelled) setLoading(false)
          return
        }
      }

      // Prefer the freshly fetched profile so edits made elsewhere (or a
      // rotated avatar URL) are reflected, but keep the cached copy if the
      // request fails rather than signing the user out.
      let profile = cached
      try {
        profile = await api.getMe()
        saveUser(profile)
      } catch {
        // Fall back to the cached user.
      }

      if (!cancelled) {
        setUser(profile)
        setLoading(false)
      }
    }

    void boot()
    return () => {
      cancelled = true
    }
  }, [])

  const login = useCallback(async (username: string, password: string) => {
    setError('')
    const res = await api.login(username, password)
    saveUser(res.user)
    setUser(res.user)
  }, [])

  const register = useCallback(async (username: string, password: string) => {
    setError('')
    const res = await api.register(username, password)
    saveUser(res.user)
    setUser(res.user)
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.logout()
    } finally {
      clearUser()
      setUser(null)
    }
  }, [])

  const applyUser = useCallback((updated: User) => {
    saveUser(updated)
    setUser(updated)
  }, [])

  const updateProfile = useCallback(
    async (patch: UpdateProfileInput) => {
      applyUser(await api.updateProfile(patch))
    },
    [applyUser],
  )

  const updateAvatar = useCallback(
    async (file: File) => {
      applyUser(await api.updateAvatar(file))
    },
    [applyUser],
  )

  const value = useMemo(
    () => ({
      user,
      loading,
      error,
      login,
      register,
      logout,
      updateProfile,
      updateAvatar,
    }),
    [
      user,
      loading,
      error,
      login,
      register,
      logout,
      updateProfile,
      updateAvatar,
    ],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return ctx
}
