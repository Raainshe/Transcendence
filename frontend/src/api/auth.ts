import { apiFetch } from '@/api/client'
import type {
  AuthResponse,
  MeResponse,
  RefreshResponse,
  LoginResult,
  VerifyLoginResponse,
  MessageResponse,
} from '@/types/api'

export interface LoginPayload {
  email: string
  password: string
}

export interface RegisterPayload {
  username: string
  email: string
  password: string
}

export function login(payload: LoginPayload): Promise<LoginResult> {
  return apiFetch<LoginResult>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
    auth: false,
  })
}

export function verifyLogin(pendingToken: string, code: string): Promise<VerifyLoginResponse> {
  return apiFetch<VerifyLoginResponse>('/auth/2fa/verify-login', {
    method: 'POST',
    body: JSON.stringify({ pending_token: pendingToken, code }),
    auth: false,
  })
}

export function setup2FA(): Promise<MessageResponse> {
  return apiFetch<MessageResponse>('/auth/2fa/setup', { method: 'POST', auth: true })
}

export function verify2FA(code: string): Promise<MessageResponse> {
  return apiFetch<MessageResponse>('/auth/2fa/verify', {
    method: 'POST',
    body: JSON.stringify({ code }),
    auth: true,
  })
}

export function disable2FA(): Promise<MessageResponse> {
  return apiFetch<MessageResponse>('/auth/2fa', { method: 'DELETE', auth: true })
}

export function register(payload: RegisterPayload): Promise<AuthResponse> {
  return apiFetch<AuthResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
    auth: false,
  })
}

export function refreshToken(): Promise<RefreshResponse> {
  return apiFetch<RefreshResponse>('/auth/refresh', {
    method: 'POST',
    auth: true,
  })
}

export function logout(): Promise<void> {
  return apiFetch<void>('/auth/logout', {
    method: 'POST',
    auth: true,
  })
}

export function getMe(): Promise<MeResponse> {
  return apiFetch<MeResponse>('/users/me', {
    method: 'GET',
    auth: true,
  })
}
