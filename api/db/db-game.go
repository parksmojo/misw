package db

import (
	"context"

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
