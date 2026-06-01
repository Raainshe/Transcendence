import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import * as authApi from '@/api/auth'
import * as usersApi from '@/api/users'
import { shouldRefreshNow } from '@/auth/jwt'
import { scheduleTokenRefresh, stopTokenRefreshTimer } from '@/auth/tokenRefresh'
import { ApiError, setBearerToken } from '@/api/client'
import type { User } from '@/types/api'

const STORAGE_KEY = 'transcendence_auth'

type StoredAuth = {
  token: string
  user: User
}

type AuthStatus = 'idle' | 'loading' | 'authenticated'

function loadStoredAuth(): StoredAuth | null {
  if (typeof localStorage === 'undefined') return null
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as StoredAuth
    if (!parsed.token || !parsed.user?.id) return null
    return parsed
  } catch {
    return null
  }
}

function saveStoredAuth(data: StoredAuth | null): void {
  if (typeof localStorage === 'undefined') return
  try {
    if (!data) {
      localStorage.removeItem(STORAGE_KEY)
      return
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
  } catch {
    /* quota / private mode */
  }
}

function applySession(token: string | null, user: User | null): void {
  setBearerToken(token)
  saveStoredAuth(token && user ? { token, user } : null)
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const user = ref<User | null>(null)
  const status = ref<AuthStatus>('idle')

  let refreshInFlight: Promise<void> | null = null
  let sessionRestore: Promise<void> | null = null

  const isAuthenticated = computed(() => status.value === 'authenticated' && !!token.value)

  function armTokenRefresh(nextToken: string): void {
    scheduleTokenRefresh(nextToken, () => refreshAccessToken())
  }

  function setSession(nextToken: string, nextUser: User): void {
    token.value = nextToken
    user.value = nextUser
    status.value = 'authenticated'
    applySession(nextToken, nextUser)
    armTokenRefresh(nextToken)
  }

  function applyUser(nextUser: User): void {
    if (!token.value) return
    user.value = nextUser
    applySession(token.value, nextUser)
  }

  function clearSession(): void {
    stopTokenRefreshTimer()
    token.value = null
    user.value = null
    status.value = 'idle'
    applySession(null, null)
  }

  async function refreshAccessToken(): Promise<void> {
    if (!token.value) return
    if (refreshInFlight) return refreshInFlight

    refreshInFlight = (async () => {
      const currentUser = user.value
      if (!currentUser) return

      try {
        const { token: nextToken } = await authApi.refreshToken()
        token.value = nextToken
        applySession(nextToken, currentUser)
        armTokenRefresh(nextToken)
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) {
          clearSession()
          return
        }
        console.warn('Token refresh failed', error)
      }
    })().finally(() => {
      refreshInFlight = null
    })

    return refreshInFlight
  }

  async function checkAndRefreshToken(): Promise<void> {
    const current = token.value
    if (!current || status.value !== 'authenticated') return
    if (shouldRefreshNow(current)) {
      await refreshAccessToken()
    } else {
      armTokenRefresh(current)
    }
  }

  async function refreshMe(): Promise<void> {
    const { user: freshUser } = await authApi.getMe()
    applyUser(freshUser)
  }

  async function login(email: string, password: string): Promise<void> {
    status.value = 'loading'
    try {
      const { user: nextUser, token: nextToken } = await authApi.login({ email, password })
      setSession(nextToken, nextUser)
    } catch (error) {
      status.value = token.value ? 'authenticated' : 'idle'
      throw error
    }
  }

  async function register(username: string, email: string, password: string): Promise<void> {
    status.value = 'loading'
    try {
      const { user: nextUser, token: nextToken } = await authApi.register({
        username,
        email,
        password,
      })
      setSession(nextToken, nextUser)
    } catch (error) {
      status.value = token.value ? 'authenticated' : 'idle'
      throw error
    }
  }

  async function logout(): Promise<void> {
    const hadToken = !!token.value
    try {
      if (hadToken) {
        await authApi.logout()
      }
    } catch {
      /* discard token locally even if server call fails */
    } finally {
      clearSession()
    }
  }

  async function updateProfile(username: string): Promise<void> {
    const { user: nextUser } = await usersApi.updateMe({ username })
    applyUser(nextUser)
  }

  async function uploadAvatar(file: File): Promise<void> {
    const { user: nextUser } = await usersApi.uploadAvatar(file)
    applyUser(nextUser)
  }

  async function removeAvatar(): Promise<void> {
    await usersApi.deleteAvatar()
    await refreshMe()
  }

  async function deleteAccount(): Promise<void> {
    await usersApi.deleteMe()
    clearSession()
  }

  async function hydrate(): Promise<void> {
    const stored = loadStoredAuth()
    if (!stored) {
      clearSession()
      return
    }

    token.value = stored.token
    user.value = stored.user
    setBearerToken(stored.token)
    status.value = 'loading'

    let activeToken = stored.token

    try {
      if (shouldRefreshNow(stored.token)) {
        const { token: refreshed } = await authApi.refreshToken()
        activeToken = refreshed
        token.value = refreshed
        setBearerToken(refreshed)
        saveStoredAuth({ token: refreshed, user: stored.user })
      }

      const { user: freshUser } = await authApi.getMe()
      setSession(activeToken, freshUser)
    } catch (error) {
      if (error instanceof ApiError && (error.status === 401 || error.status === 404)) {
        clearSession()
        return
      }
      status.value = 'authenticated'
      armTokenRefresh(activeToken)
    }
  }

  /** Resolves after the initial session restore attempt (success or guest). */
  async function whenReady(): Promise<void> {
    if (status.value === 'authenticated') return
    if (status.value === 'idle' && !loadStoredAuth()) return
    if (!sessionRestore) {
      sessionRestore = hydrate().finally(() => {
        sessionRestore = null
      })
    }
    await sessionRestore
  }

  return {
    token,
    user,
    status,
    isAuthenticated,
    login,
    register,
    logout,
    hydrate,
    whenReady,
    refreshMe,
    refreshAccessToken,
    checkAndRefreshToken,
    updateProfile,
    uploadAvatar,
    removeAvatar,
    deleteAccount,
    clearSession,
  }
})
