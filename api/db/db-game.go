package db

import (
	"context"
	"errors"
	"misw-api/model"

	"github.com/jackc/pgx/v5"
)

func CreateGame(conn *pgx.Conn, userID int, values [][]int, board [][]string) (int, error) {
	var gameID int
	err := conn.QueryRow(
		context.Background(),
		`INSERT INTO games (user_id, values, board)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		userID, values, board,
	).Scan(&gameID)
	if err != nil {
		return -1, err
	}
	return gameID, nil
}

func GetGame(conn *pgx.Conn, userID, gameID int) (*model.Game, error) {
	var game model.Game
	err := conn.QueryRow(
		context.Background(), 
		`SELECT id, created_at, updated_at, user_id, start_time, end_time, moves_count, values, board, won FROM games WHERE user_id=$1 AND id=$2`, 
		userID, gameID,
	).Scan(
			&game.ID,
			&game.CreatedAt,
			&game.UpdatedAt,
			&game.UserID,
			&game.StartTime,
			&game.EndTime,
			&game.MovesCount,
			&game.Values,
			&game.Board,
			&game.Won,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &game, nil
}

func EndGame(conn *pgx.Conn, userID, gameID int, board [][]string, won bool) error {
	_, err := conn.Exec(
		context.Background(),
		`UPDATE games
		 SET end_time = NOW(), won = $4, board = $1, moves_count = moves_count + 1
		 WHERE user_id = $2 AND id = $3`,
		board, userID, gameID, won,
	)
	return err
}

func UpdateGame(conn *pgx.Conn, userID, gameID int, board [][]string) error {
	_, err := conn.Exec(
		context.Background(),
		`UPDATE games
		 SET board = $1, moves_count = moves_count + 1
		 WHERE user_id = $2 AND id = $3`,
		board, userID, gameID,
	)
	return err
}

func GetGamesForUser(conn *pgx.Conn, userID int) ([]model.Game, error) {
	rows, err := conn.Query(context.Background(), "SELECT id, created_at, updated_at, user_id, start_time, end_time, moves_count, values, board, won FROM games WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []model.Game
	for rows.Next() {
		var game model.Game
		if err := rows.Scan(
			&game.ID,
			&game.CreatedAt,
			&game.UpdatedAt,
			&game.UserID,
			&game.StartTime,
			&game.EndTime,
			&game.MovesCount,
			&game.Values,
			&game.Board,
			&game.Won,
		); err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	return games, nil
}