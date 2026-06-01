export type UserRole = 'user' | 'admin'

export interface User {
  id: string
  username: string
  email: string
  avatar_url: string | null
  role: UserRole
  is_2fa_enabled: boolean
  created_at: string
  updated_at: string
  last_seen_at: string | null
  is_online?: boolean
}

export interface AuthResponse {
  user: User
  token: string
}

export interface RefreshResponse {
  token: string
}

export interface MeResponse {
  user: User
}

export interface ApiErrorBody {
  error?: string
}

export interface UpdateMePayload {
  username?: string
  avatar_url?: string
}

export interface UserStats {
  games_played: number
  wins: number
  best_score: number
  total_lines: number
  avg_score: number
}
