package endpoints

import (
	"encoding/json"
	"misw-api/auth"
	"misw-api/db"
	"net/http"
	"time"
)

type unfinishedGame struct {
	ID         int        `json:"id"`
	Board      [][]string `json:"board"`
	MovesCount int        `json:"movesCount"`
	CreatedAt  string     `json:"createdAt"`
	UpdatedAt  string     `json:"updatedAt"`
}

func GetUnfinishedGamesHandler(w http.ResponseWriter, r *http.Request) {
	conn := db.OpenConnection()
	defer db.CloseConnection(conn)

	user := auth.ValidateRequestingUser(w, r, conn)
	if user == nil {
		return
	}

	games, err := db.GetUnfinishedGamesForUser(conn, user.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve games", http.StatusInternalServerError)
		return
	}

	response := make([]unfinishedGame, 0, len(games))
	for _, game := range games {
		response = append(response, unfinishedGame{
			ID:         game.ID,
			Board:      game.Board,
			MovesCount: game.MovesCount,
			CreatedAt:  game.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  game.UpdatedAt.Format(time.RFC3339),
		})
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
