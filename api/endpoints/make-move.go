package endpoints

import (
	"encoding/json"
	"misw-api/auth"
	"misw-api/db"
	"net/http"
	"strconv"
)

type makeMoveRequest struct {
	GameID int `json:"gameId"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

type makeMoveResponse struct {
	Board  [][]string `json:"board"`
	Result *bool      `json:"result,omitempty"`
}

func MakeMoveHandler(w http.ResponseWriter, r *http.Request) {
	conn := db.OpenConnection()
	defer db.CloseConnection(conn)

	user := auth.ValidateRequestingUser(w, r, conn)
	if user == nil {
		return
	}

	var body makeMoveRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	game, err := db.GetGame(conn, user.ID, body.GameID)
	if err != nil {
		http.Error(w, "Could not get game", http.StatusInternalServerError)
		return
	}
	if game == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}
	if game.EndTime != nil {
		http.Error(w, "Game has ended", http.StatusBadRequest)
		return
	}

	if body.X < 0 || body.Y < 0 || body.Y >= len(game.Board) || body.X >= len(game.Board[0]) {
		http.Error(w, "Invalid request: Coordinate out of bounds", http.StatusBadRequest)
		return
	}
	if game.Board[body.Y][body.X] != " " {
		http.Error(w, "Invalid request: Coordinate already revealed", http.StatusBadRequest)
		return
	}

	if game.Values[body.Y][body.X] == 9 {
		for y, row := range game.Values {
			for x, val := range row {
				if val != 9 {
					continue
				}
				if x == body.X && y == body.Y {
					game.Board[y][x] = "X"
				} else {
					game.Board[y][x] = "B"
				}
			}
		}
		err := db.EndGame(conn, user.ID, game.ID, game.Board, false)
		if err != nil {
			http.Error(w, "Couldn't process move", http.StatusInternalServerError)
			return
		}
		result := false
		jsonBytes, err := json.Marshal(makeMoveResponse{Board: game.Board, Result: &result})
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
		return
	}

	height := len(game.Board)
	width := len(game.Board[0])
	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}
	queue := make([][2]int, 1, width*height)
	queue[0] = [2]int{body.X, body.Y}
	for head := 0; head < len(queue); head++ {
		x, y := queue[head][0], queue[head][1]
		if visited[y][x] {
			continue
		}
		visited[y][x] = true
		if game.Board[y][x] != " " {
			continue
		}
		val := game.Values[y][x]
		game.Board[y][x] = strconv.Itoa(val)
		if val != 0 {
			continue
		}
		for ny := y - 1; ny <= y+1; ny++ {
			if ny < 0 || ny >= height {
				continue
			}
			for nx := x - 1; nx <= x+1; nx++ {
				if nx < 0 || nx >= width || (nx == x && ny == y) {
					continue
				}
				if visited[ny][nx] || game.Board[ny][nx] != " " {
					continue
				}
				queue = append(queue, [2]int{nx, ny})
			}
		}
	}

	gameOver := true
	for y := 0; y < height && gameOver; y++ {
		boardRow := game.Board[y]
		valueRow := game.Values[y]
		for x := 0; x < width; x++ {
			if boardRow[x] == " " && valueRow[x] != 9 {
				gameOver = false
				break
			}
		}
	}

	if gameOver {
		err = db.EndGame(conn, user.ID, game.ID, game.Board, true)
		if err != nil {
			http.Error(w, "Couldn't process move", http.StatusInternalServerError)
			return
		}
		result := true
		jsonBytes, err := json.Marshal(makeMoveResponse{Board: game.Board, Result: &result})
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
		return
	}

	err = db.UpdateGame(conn, user.ID, game.ID, game.Board)
	if err != nil {
		http.Error(w, "Couldn't process move", http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(makeMoveResponse{Board: game.Board})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
