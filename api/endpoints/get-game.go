package endpoints

import (
	"encoding/json"
	"misw-api/auth"
	"misw-api/db"
	"net/http"
)

type getGameRequest struct {
	ID int `json:"id"`
}

type getGameResponse struct {
	Board [][]string `json:"board"`
}

func GetGameHandler(w http.ResponseWriter, r *http.Request) {
	conn := db.OpenConnection()
	defer db.CloseConnection(conn)

	user := auth.ValidateRequestingUser(w, r, conn)
	if user == nil {
		return
	}

	var body getGameRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	game, err := db.GetGame(conn, user.ID, body.ID)
	if err != nil {
		http.Error(w, "Could not get game", http.StatusInternalServerError)
		return
	}
	if game == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(getGameResponse{ Board: game.Board })
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}