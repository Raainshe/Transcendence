import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import * as authApi from '@/api/auth'
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

  const isAuthenticated = computed(() => status.value === 'authenticated' && !!token.value)

  function setSession(nextToken: string, nextUser: User): void {
    token.value = nextToken
    user.value = nextUser
    status.value = 'authenticated'
    applySession(nextToken, nextUser)
  }

  function clearSession(): void {
    token.value = null
    user.value = null
    status.value = 'idle'
    applySession(null, null)
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

    try {
      const { user: freshUser } = await authApi.getMe()
      setSession(stored.token, freshUser)
    } catch (error) {
      if (error instanceof ApiError && (error.status === 401 || error.status === 404)) {
        clearSession()
        return
      }
      status.value = 'authenticated'
    }
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
    clearSession,
  }
})
