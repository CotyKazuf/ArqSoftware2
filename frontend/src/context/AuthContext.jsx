import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { getMe, login as loginRequest } from '../services/authService'

const AuthContext = createContext(null)
const TOKEN_STORAGE_KEY = 'lokis-perfume-token'
const USER_STORAGE_KEY = 'lokis-perfume-user'

const readStoredUser = () => {
  if (typeof window === 'undefined') return null
  try {
    const saved = window.localStorage.getItem(USER_STORAGE_KEY)
    return saved ? JSON.parse(saved) : null
  } catch {
    return null
  }
}

const readStoredToken = () => {
  if (typeof window === 'undefined') return null
  try {
    return window.localStorage.getItem(TOKEN_STORAGE_KEY)
  } catch {
    return null
  }
}

const persistSession = (token, user) => {
  if (typeof window === 'undefined') return
  try {
    if (token) {
      window.localStorage.setItem(TOKEN_STORAGE_KEY, token)
    } else {
      window.localStorage.removeItem(TOKEN_STORAGE_KEY)
    }
    if (user) {
      window.localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user))
    } else {
      window.localStorage.removeItem(USER_STORAGE_KEY)
    }
  } catch {
    // Ignore storage quota errors
  }
}

export function AuthProvider({ children }) {
  const [user, setUser] = useState(() => readStoredUser())
  const [token, setToken] = useState(() => readStoredToken())
  const [isLoading, setIsLoading] = useState(true)
  const [sessionError, setSessionError] = useState('')

  const clearSession = useCallback(() => {
    setUser(null)
    setToken(null)
    setSessionError('')
    persistSession(null, null)
  }, [])

  const loadSessionFromStorage = useCallback(async () => {
    setIsLoading(true)
    const storedToken = readStoredToken()
    const storedUser = readStoredUser()
    if (!storedToken) {
      clearSession()
      setSessionError('')
      setIsLoading(false)
      return
    }
    setToken(storedToken)
    if (storedUser) {
      setUser(storedUser)
    }

    try {
      const profile = await getMe(storedToken)
      setUser(profile)
      setSessionError('')
      persistSession(storedToken, profile)
    } catch (error) {
      console.error('auth: restoring session failed', error)
      clearSession()
      const message =
        error.code === 'AUTHENTICATION_FAILED'
          ? 'Tu sesión expiró. Iniciá sesión nuevamente.'
          : 'No pudimos validar tu sesión.'
      setSessionError(message)
    } finally {
      setIsLoading(false)
    }
  }, [clearSession])

  useEffect(() => {
    loadSessionFromStorage()
  }, [loadSessionFromStorage])

  const login = useCallback(
    async ({ email, password }) => {
      const result = await loginRequest({ email, password })
      setUser(result.user)
      setToken(result.token)
      setSessionError('')
      persistSession(result.token, result.user)
      return result.user
    },
    [],
  )

  const logout = useCallback(() => {
    clearSession()
  }, [clearSession])

  const value = useMemo(
    () => ({
      user,
      token,
      sessionError,
      isLoading,
      isAuthenticated: Boolean(user && token),
      login,
      logout,
      loadSessionFromStorage,
    }),
    [isLoading, loadSessionFromStorage, login, logout, sessionError, token, user],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth debe usarse dentro de AuthProvider')
  }
  return context
}
