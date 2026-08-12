package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/model"
)

var (
	ErrLobbyFull       = errors.New("lobby is full")
	ErrAlreadyInLobby  = errors.New("user is already in this lobby")
	ErrNotLobbyMember  = errors.New("user is not a lobby member")
	ErrLobbyNotWaiting = errors.New("lobby is not accepting changes")
)

type LobbyRepository interface {
	Create(ctx context.Context, lobby *model.Lobby, hostUserID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Lobby, error)
	FindByInviteCode(ctx context.Context, code string) (*model.Lobby, error)
	FindDetail(ctx context.Context, id uuid.UUID) (*model.LobbyDetail, error)
	FindWaitingLobbyByUser(ctx context.Context, userID uuid.UUID) (*model.Lobby, error)
	AddMember(ctx context.Context, lobbyID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, lobbyID, userID uuid.UUID) error
	SetReady(ctx context.Context, lobbyID, userID uuid.UUID, ready bool) error
	DeleteLobby(ctx context.Context, id uuid.UUID) error
	LinkGame(ctx context.Context, lobbyID, gameID uuid.UUID, seed int64) error
	MemberCount(ctx context.Context, lobbyID uuid.UUID) (int, error)
	IsMember(ctx context.Context, lobbyID, userID uuid.UUID) (bool, error)
	FindByGameID(ctx context.Context, gameID uuid.UUID) (*model.Lobby, error)
	UpdateLobbyName(ctx context.Context, lobbyID uuid.UUID, name string) error
}

type lobbyRepository struct {
	db *sql.DB
}

func NewLobbyRepository(db *sql.DB) LobbyRepository {
	return &lobbyRepository{db: db}
}

func (r *lobbyRepository) Create(ctx context.Context, lobby *model.Lobby, hostUserID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const insertLobby = `
		INSERT INTO lobbies (id, host_user_id, invite_code, max_players, status, created_at, name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if _, err := tx.ExecContext(ctx, insertLobby,
		lobby.ID.String(), lobby.HostUserID.String(), lobby.InviteCode,
		lobby.MaxPlayers, lobby.Status, lobby.CreatedAt, lobby.Name,
	); err != nil {
		return err
	}

	const insertMember = `
		INSERT INTO lobby_members (id, lobby_id, user_id, is_ready, joined_at)
		VALUES ($1, $2, $3, false, $4)
	`
	if _, err := tx.ExecContext(ctx, insertMember,
		uuid.New().String(), lobby.ID.String(), hostUserID.String(), time.Now().UTC(),
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *lobbyRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Lobby, error) {
	const q = `
		SELECT id, host_user_id, invite_code, max_players, status, game_id, shared_seed, created_at, name
		FROM lobbies WHERE id = $1
	`
	return r.scanLobby(r.db.QueryRowContext(ctx, q, id.String()))
}

func (r *lobbyRepository) FindByGameID(ctx context.Context, gameID uuid.UUID) (*model.Lobby, error) {
	const q = `
		SELECT id, host_user_id, invite_code, max_players, status, game_id, shared_seed, created_at, name
		FROM lobbies WHERE game_id = $1
	`
	return r.scanLobby(r.db.QueryRowContext(ctx, q, gameID.String()))
}

func (r *lobbyRepository) FindByInviteCode(ctx context.Context, code string) (*model.Lobby, error) {
	const q = `
		SELECT id, host_user_id, invite_code, max_players, status, game_id, shared_seed, created_at, name
		FROM lobbies WHERE invite_code = $1
	`
	return r.scanLobby(r.db.QueryRowContext(ctx, q, strings.ToUpper(strings.TrimSpace(code))))
}

func (r *lobbyRepository) FindWaitingLobbyByUser(ctx context.Context, userID uuid.UUID) (*model.Lobby, error) {
	const q = `
		SELECT l.id, l.host_user_id, l.invite_code, l.max_players, l.status, l.game_id, l.shared_seed, l.created_at, l.name
		FROM lobbies l
		JOIN lobby_members m ON m.lobby_id = l.id
		WHERE m.user_id = $1 AND l.status = 'waiting'
		LIMIT 1
	`
	return r.scanLobby(r.db.QueryRowContext(ctx, q, userID.String()))
}

func (r *lobbyRepository) FindDetail(ctx context.Context, id uuid.UUID) (*model.LobbyDetail, error) {
	lobby, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	const q = `
		SELECT u.id, u.username, u.avatar_url, m.is_ready, m.joined_at
		FROM lobby_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.lobby_id = $1
		ORDER BY m.joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, q, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.LobbyMemberView
	for rows.Next() {
		var m model.LobbyMemberView
		var userIDStr string
		if err := rows.Scan(&userIDStr, &m.Username, &m.AvatarURL, &m.IsReady, &m.JoinedAt); err != nil {
			return nil, err
		}
		m.UserID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if members == nil {
		members = []model.LobbyMemberView{}
	}

	return &model.LobbyDetail{Lobby: *lobby, Members: members}, nil
}

func (r *lobbyRepository) AddMember(ctx context.Context, lobbyID, userID uuid.UUID) error {
	const q = `
		INSERT INTO lobby_members (id, lobby_id, user_id, is_ready, joined_at)
		VALUES ($1, $2, $3, false, now())
	`
	_, err := r.db.ExecContext(ctx, q, uuid.New().String(), lobbyID.String(), userID.String())
	return err
}

func (r *lobbyRepository) RemoveMember(ctx context.Context, lobbyID, userID uuid.UUID) error {
	const q = `DELETE FROM lobby_members WHERE lobby_id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, q, lobbyID.String(), userID.String())
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotLobbyMember
	}
	return nil
}

func (r *lobbyRepository) SetReady(ctx context.Context, lobbyID, userID uuid.UUID, ready bool) error {
	const q = `
		UPDATE lobby_members SET is_ready = $3
		WHERE lobby_id = $1 AND user_id = $2
	`
	res, err := r.db.ExecContext(ctx, q, lobbyID.String(), userID.String(), ready)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotLobbyMember
	}
	return nil
}

func (r *lobbyRepository) DeleteLobby(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM lobbies WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, id.String())
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *lobbyRepository) LinkGame(ctx context.Context, lobbyID, gameID uuid.UUID, seed int64) error {
	const q = `
		UPDATE lobbies
		SET game_id = $2, shared_seed = $3, status = 'closed'
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, q, lobbyID.String(), gameID.String(), seed)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *lobbyRepository) MemberCount(ctx context.Context, lobbyID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM lobby_members WHERE lobby_id = $1`
	var n int
	err := r.db.QueryRowContext(ctx, q, lobbyID.String()).Scan(&n)
	return n, err
}

func (r *lobbyRepository) IsMember(ctx context.Context, lobbyID, userID uuid.UUID) (bool, error) {
	const q = `SELECT 1 FROM lobby_members WHERE lobby_id = $1 AND user_id = $2 LIMIT 1`
	var one int
	err := r.db.QueryRowContext(ctx, q, lobbyID.String(), userID.String()).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *lobbyRepository) scanLobby(row *sql.Row) (*model.Lobby, error) {
	var l model.Lobby
	var idStr, hostStr string
	var gameID sql.NullString
	var sharedSeed sql.NullInt64
	err := row.Scan(
		&idStr, &hostStr, &l.InviteCode, &l.MaxPlayers, &l.Status,
		&gameID, &sharedSeed, &l.CreatedAt, &l.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	l.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	l.HostUserID, err = uuid.Parse(hostStr)
	if err != nil {
		return nil, err
	}
	if gameID.Valid {
		gid, err := uuid.Parse(gameID.String)
		if err != nil {
			return nil, err
		}
		l.GameID = &gid
	}
	if sharedSeed.Valid {
		v := sharedSeed.Int64
		l.SharedSeed = &v
	}
	return &l, nil
}

func (r *lobbyRepository) UpdateLobbyName(ctx context.Context, lobbyID uuid.UUID, name string) error {
	const q = `UPDATE lobbies SET name = $2 WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, lobbyID.String(), name)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
