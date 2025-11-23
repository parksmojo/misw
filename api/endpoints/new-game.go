package endpoints

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"misw-api/auth"
	"misw-api/db"
	"misw-api/model"
	"net/http"
	"time"
)

type newGameRequest struct {
	Width     int `json:"width"`
	Height    int `json:"height"`
	BombCount int `json:"bombCount"`
}

type newGameResponse struct {
	ID    int        `json:"id"`
	Board [][]string `json:"board"`
}

func NewGameHandler(w http.ResponseWriter, r *http.Request) {
	conn := db.OpenConnection()
	defer db.CloseConnection(conn)

	user := auth.ValidateRequestingUser(w, r, conn)
	if user == nil {
		return
	}

	var body newGameRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Width < 1 || body.Width > 100 || body.Height < 1 || body.Height > 100 || body.BombCount < 1 || body.BombCount > body.Width*body.Height {
		http.Error(w, "Cannot create specified game board", http.StatusBadRequest)
		return
	}

	possibleLocations := make([]model.Coord, 0, body.Width*body.Height)
	for i := 0; i < body.Height; i++ {
		for j := 0; j < body.Width; j++ {
			possibleLocations = append(possibleLocations, model.Coord{X: j, Y: i})
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(possibleLocations), func(i, j int) {
		possibleLocations[i], possibleLocations[j] = possibleLocations[j], possibleLocations[i]
	})
	possibleLocations = possibleLocations[:body.BombCount]

	values := make([][]int, body.Height)
	for i := range values {
		values[i] = make([]int, body.Width)
		for j := range values[i] {
			values[i][j] = 0
		}
	}

	for _, coord := range possibleLocations {
		y, x := coord.Y, coord.X
		values[y][x] = 9

		for ny := y - 1; ny <= y+1; ny++ {
			if ny < 0 || ny >= body.Height {
				continue
			}
			row := values[ny]
			for nx := x - 1; nx <= x+1; nx++ {
				if nx < 0 || nx >= body.Width || (ny == y && nx == x) {
					continue
				}
				if row[nx] != 9 {
					row[nx]++
				}
			}
		}
	}

	board := make([][]string, body.Height)
	for i := range board {
		board[i] = make([]string, body.Width)
		for j := range board[i] {
			board[i][j] = " "
		}
	}

	gameId, err := db.CreateGame(conn, user.ID, values, board)
	if err != nil {
		fmt.Println("Error creating game:", err)
		http.Error(w, "Could not create game", http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(newGameResponse{ID: gameId, Board: board})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
