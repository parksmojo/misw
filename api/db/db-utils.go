package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

func OpenConnection() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic("Unable to connect to database: " + err.Error() + "\n")
	}
	return conn
}

func CloseConnection(conn *pgx.Conn) {
	conn.Close(context.Background())
}
