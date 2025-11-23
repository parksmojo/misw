package db

import (
	"context"
	"fmt"
	"misw-api/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func GetUser(conn *pgx.Conn, username, password string) (*model.User, error) {
	row := conn.QueryRow(context.Background(),
		`SELECT id, created_at, updated_at, username, password, last_login FROM users WHERE username = $1`, username)

	var user model.User
	err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Username, &user.Password, &user.LastLogin)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

func CreateUser(conn *pgx.Conn, username, password string) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err.Error())
		return nil, fmt.Errorf("failed to create user");
	}

	row := conn.QueryRow(context.Background(),
		`INSERT INTO users (username, password) VALUES ($1, $2)
		 RETURNING id, created_at, updated_at, username, password, last_login`,
		username, string(hashedPassword),
	)

	var user model.User
	err = row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Username, &user.Password, &user.LastLogin)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fmt.Errorf("username already exists")
		}
		return nil, err
	}

	return &user, nil
}