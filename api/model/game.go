package model

import "time"

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Game struct {
	ID              int         `json:"id"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	UserID          int         `json:"user_id"`
	StartTime       time.Time   `json:"start_time"`
	EndTime         *time.Time  `json:"end_time,omitempty"`
	Width           int         `json:"width"`
	Height          int         `json:"height"`
	NumberOfBombs   int         `json:"number_of_bombs"`
	NumberOfMoves   int         `json:"number_of_moves"`
	BombLocations   []Coord     `json:"bomb_locations"`
	Board           [][]string  `json:"board"`
	Won             *bool       `json:"won"`
}
