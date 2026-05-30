package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"backend/internal/model"
)

type GameRepository interface {
	RecordMatch(ctx context.Context, game *model.Game, player *model.GamePlayer) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error)
	ListLeaderboard(ctx context.Context, limit int) ([]model.LeaderboardEntry, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*model.UserStats, error)
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

func (r *gameRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*model.UserStats, error) {
	const q = `
		SELECT
			COUNT(*)                               AS games_played,
			SUM(CASE WHEN is_winner THEN 1 ELSE 0 END) AS wins,
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
