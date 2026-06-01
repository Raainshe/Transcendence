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

export interface UsersListResponse {
  users: User[]
  total: number
}

export interface FriendsResponse {
  friends: User[]
}

export interface FriendRequestsResponse {
  requests: User[]
}

export interface BlockedUsersResponse {
  blocked: User[]
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

export interface CreateGamePayload {
  mode: string
  score: number
  lines_cleared: number
  level_reached: number
  started_at: string
  finished_at: string
  is_winner: boolean
}

export interface LeaderboardEntry {
  rank: number
  user_id: string
  username: string
  avatar_url: string | null
  score: number
  lines_cleared: number
  level_reached: number
  mode: string
  finished_at: string | null
}

export interface CreateGameResponse {
  game: {
    id: string
    mode: string
    status: string
    created_at: string
    finished_at: string | null
  }
}
