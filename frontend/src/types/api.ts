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

export interface TwoFactorRequiredResponse {
  two_factor_required: true
  pending_token: string
}

export type LoginResult = AuthResponse | TwoFactorRequiredResponse

export interface VerifyLoginResponse {
  token: string
}

export interface MessageResponse {
  message: string
}

export interface RefreshResponse {
  token: string
}

export interface MeResponse {
  user: User
}

export interface UserResponse {
  user: User
}

export interface GameSummary {
  id: string
  mode: string
  status: string
  created_at: string
  finished_at: string | null
}

export interface GamePlayer {
  id: string
  game_id: string
  user_id: string
  score: number
  lines_cleared: number
  level_reached: number
  placement: number | null
  is_winner: boolean
}

export interface GameDetail extends GameSummary {
  players: GamePlayer[]
}

export interface GamesListResponse {
  games: GameSummary[]
  total: number
}

export interface GameDetailResponse {
  game: GameDetail
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

export interface LobbyMember {
  user_id: string
  username: string
  avatar_url: string | null
  is_ready: boolean
  joined_at: string
}

export interface LobbyDetail {
  id: string
  host_user_id: string
  invite_code: string
  max_players: number
  status: 'waiting' | 'closed'
  game_id: string | null
  shared_seed: number | null
  created_at: string
  members: LobbyMember[]
}

export interface LobbyResponse {
  lobby: LobbyDetail
}

export interface CreateLobbyPayload {
  max_players: number
}

export interface JoinLobbyByCodePayload {
  invite_code: string
}

export interface SetReadyPayload {
  ready: boolean
}

export interface MatchStartPlayer {
  user_id: string
  username: string
  avatar_url: string | null
}

export interface StartLobbyResult {
  game_id: string
  shared_seed: number
  players: MatchStartPlayer[]
}

export interface MatchPlayerView {
  user_id: string
  username: string
  avatar_url: string | null
}

export interface MatchDetail {
  game_id: string
  status: string
  mode: string
  shared_seed: number
  players: MatchPlayerView[]
  results?: MatchEndedPayload
}

export interface PlayerConnectionBroadcast {
  user_id: string
}

export interface MatchResponse {
  match: MatchDetail | null
}

export interface PlayerStateUpload {
  score: number
  lines: number
  level: number
  alive: boolean
  board: string
}

export interface PlayerStateBroadcast extends PlayerStateUpload {
  user_id: string
}

export interface PlayerEliminatedUpload {
  reason: string
  score: number
  lines: number
  level: number
}

export interface PlayerEliminatedBroadcast {
  user_id: string
  reason: string
  placement: number
}

export interface MatchEndedPlayer {
  user_id: string
  username: string
  score: number
  lines: number
  level: number
  placement: number
  is_winner: boolean
  elimination_reason?: string | null
}

export interface MatchEndedPayload {
  winner_user_id?: string | null
  players: MatchEndedPlayer[]
}
