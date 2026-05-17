package repository

import "database/sql"

type GameRepository interface{}

type gameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) GameRepository {
	return &gameRepository{db: db}
}
