package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"backend/internal/model"
)

type GameRepository interface {
	RecordMatch(ctx context.Context, game *model.Game, player *model.GamePlayer) error
	CreateMultiplayerMatch(ctx context.Context, game *model.Game, players []model.GamePlayer) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error)
	ListGames(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Game, error)
	CountGames(ctx context.Context, userID *uuid.UUID) (int, error)
	FindGameDetail(ctx context.Context, id uuid.UUID) (*model.GameDetail, error)
	ListLeaderboard(ctx context.Context, limit int) ([]model.LeaderboardEntry, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*model.UserStats, error)
	IsGamePlayer(ctx context.Context, gameID, userID uuid.UUID) (bool, error)
	ListMatchPlayers(ctx context.Context, gameID uuid.UUID) ([]model.MatchPlayerView, error)
	FinishMultiplayerMatch(ctx context.Context, gameID uuid.UUID, finishedAt time.Time, players []model.GamePlayer) error
}

type gameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) RecordMatch(ctx context.Context, game *model.Game, player *model.GamePlayer) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const insertGame = `
		INSERT INTO games (id, mode, status, created_at, finished_at)
		VALUES ($1, $2, 'finished', $3, $4)
	`
	if _, err := tx.ExecContext(ctx, insertGame,
		game.ID.String(), game.Mode, game.CreatedAt, game.FinishedAt,
	); err != nil {
		return err
	}

	const insertPlayer = `
		INSERT INTO game_players (id, game_id, user_id, score, lines_cleared, level_reached, placement, is_winner)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err := tx.ExecContext(ctx, insertPlayer,
		player.ID.String(), player.GameID.String(), player.UserID.String(),
		player.Score, player.LinesCleared, player.LevelReached, player.Placement, player.IsWinner,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *gameRepository) CreateMultiplayerMatch(ctx context.Context, game *model.Game, players []model.GamePlayer) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const insertGame = `
		INSERT INTO games (id, mode, status, created_at, finished_at)
		VALUES ($1, $2, $3, $4, NULL)
	`
	if _, err := tx.ExecContext(ctx, insertGame,
		game.ID.String(), game.Mode, game.Status, game.CreatedAt,
	); err != nil {
		return err
	}

	const insertPlayer = `
		INSERT INTO game_players (id, game_id, user_id, score, lines_cleared, level_reached, placement, is_winner)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	for i := range players {
		p := &players[i]
		if _, err := tx.ExecContext(ctx, insertPlayer,
			p.ID.String(), p.GameID.String(), p.UserID.String(),
			p.Score, p.LinesCleared, p.LevelReached, p.Placement, p.IsWinner,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *gameRepository) FinishMultiplayerMatch(
	ctx context.Context,
	gameID uuid.UUID,
	finishedAt time.Time,
	players []model.GamePlayer,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const updateGame = `
		UPDATE games
		SET status = 'finished', finished_at = $1
		WHERE id = $2 AND status = 'in_progress'
	`
	res, err := tx.ExecContext(ctx, updateGame, finishedAt, gameID.String())
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

	const updatePlayer = `
		UPDATE game_players
		SET score = $1, lines_cleared = $2, level_reached = $3, placement = $4, is_winner = $5
		WHERE game_id = $6 AND user_id = $7
	`
	for i := range players {
		p := &players[i]
		if _, err := tx.ExecContext(ctx, updatePlayer,
			p.Score, p.LinesCleared, p.LevelReached, p.Placement, p.IsWinner,
			gameID.String(), p.UserID.String(),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *gameRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error) {
	const q = `SELECT id, mode, status, created_at, finished_at FROM games WHERE id = $1`
	var g model.Game
	var idStr string
	err := r.db.QueryRowContext(ctx, q, id.String()).Scan(
		&idStr, &g.Mode, &g.Status, &g.CreatedAt, &g.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	g.ID, err = uuid.Parse(idStr)
	return &g, err
}

func (r *gameRepository) ListGames(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Game, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if userID != nil {
		const q = `
			SELECT DISTINCT g.id, g.mode, g.status, g.created_at, g.finished_at
			FROM games g JOIN game_players gp ON gp.game_id = g.id
			WHERE gp.user_id = $1
			ORDER BY g.finished_at DESC NULLS LAST, g.created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.QueryContext(ctx, q, userID.String(), limit, offset)
	} else {
		const q = `
			SELECT id, mode, status, created_at, finished_at
			FROM games
			ORDER BY finished_at DESC NULLS LAST, created_at DESC
			LIMIT $1 OFFSET $2
		`
		rows, err = r.db.QueryContext(ctx, q, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []model.Game
	for rows.Next() {
		var g model.Game
		var idStr string
		if err := rows.Scan(&idStr, &g.Mode, &g.Status, &g.CreatedAt, &g.FinishedAt); err != nil {
			return nil, err
		}
		g.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, rows.Err()
}

func (r *gameRepository) CountGames(ctx context.Context, userID *uuid.UUID) (int, error) {
	var n int
	if userID != nil {
		err := r.db.QueryRowContext(ctx,
			`SELECT COUNT(DISTINCT g.id) FROM games g JOIN game_players gp ON gp.game_id = g.id WHERE gp.user_id = $1`,
			userID.String(),
		).Scan(&n)
		return n, err
	}
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM games`).Scan(&n)
	return n, err
}

func (r *gameRepository) FindGameDetail(ctx context.Context, id uuid.UUID) (*model.GameDetail, error) {
	game, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	const q = `
		SELECT id, game_id, user_id, score, lines_cleared, level_reached, placement, is_winner
		FROM game_players WHERE game_id = $1
		ORDER BY score DESC
	`
	rows, err := r.db.QueryContext(ctx, q, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []model.GamePlayer
	for rows.Next() {
		var p model.GamePlayer
		var idStr, gameIDStr, userIDStr string
		if err := rows.Scan(&idStr, &gameIDStr, &userIDStr,
			&p.Score, &p.LinesCleared, &p.LevelReached, &p.Placement, &p.IsWinner); err != nil {
			return nil, err
		}
		if p.ID, err = uuid.Parse(idStr); err != nil {
			return nil, err
		}
		if p.GameID, err = uuid.Parse(gameIDStr); err != nil {
			return nil, err
		}
		if p.UserID, err = uuid.Parse(userIDStr); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &model.GameDetail{Game: *game, Players: players}, nil
}

func (r *gameRepository) ListLeaderboard(ctx context.Context, limit int) ([]model.LeaderboardEntry, error) {
	const q = `
		SELECT
			ROW_NUMBER() OVER (ORDER BY gp.score DESC),
			u.id, u.username, u.avatar_url,
			gp.score, gp.lines_cleared, gp.level_reached,
			g.mode, g.finished_at
		FROM game_players gp
		JOIN users u ON u.id = gp.user_id
		JOIN games g ON g.id = gp.game_id
		ORDER BY gp.score DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []model.LeaderboardEntry
	for rows.Next() {
		var e model.LeaderboardEntry
		var userIDStr string
		if err := rows.Scan(
			&e.Rank, &userIDStr, &e.Username, &e.AvatarURL,
			&e.Score, &e.LinesCleared, &e.LevelReached,
			&e.Mode, &e.FinishedAt,
		); err != nil {
			return nil, err
		}
		e.UserID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *gameRepository) IsGamePlayer(ctx context.Context, gameID, userID uuid.UUID) (bool, error) {
	const q = `SELECT 1 FROM game_players WHERE game_id = $1 AND user_id = $2 LIMIT 1`
	var one int
	err := r.db.QueryRowContext(ctx, q, gameID.String(), userID.String()).Scan(&one)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *gameRepository) ListMatchPlayers(ctx context.Context, gameID uuid.UUID) ([]model.MatchPlayerView, error) {
	const q = `
		SELECT u.id, u.username, u.avatar_url
		FROM game_players gp
		JOIN users u ON u.id = gp.user_id
		WHERE gp.game_id = $1
		ORDER BY gp.id ASC
	`
	rows, err := r.db.QueryContext(ctx, q, gameID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []model.MatchPlayerView
	for rows.Next() {
		var p model.MatchPlayerView
		var userIDStr string
		if err := rows.Scan(&userIDStr, &p.Username, &p.AvatarURL); err != nil {
			return nil, err
		}
		p.UserID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if players == nil {
		players = []model.MatchPlayerView{}
	}
	return players, nil
}

func (r *gameRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*model.UserStats, error) {
	const q = `
		SELECT
			COUNT(*)                               AS games_played,
			COALESCE(SUM(CASE WHEN is_winner THEN 1 ELSE 0 END), 0) AS wins,
			COALESCE(MAX(score), 0)                AS best_score,
			COALESCE(SUM(lines_cleared), 0)        AS total_lines,
			COALESCE(AVG(score)::int, 0)           AS avg_score
		FROM game_players
		WHERE user_id = $1
	`
	var s model.UserStats
	err := r.db.QueryRowContext(ctx, q, userID.String()).Scan(
		&s.GamesPlayed, &s.Wins, &s.BestScore, &s.TotalLines, &s.AvgScore,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
