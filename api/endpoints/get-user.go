package endpoints

import (
	"encoding/json"
	"misw-api/auth"
	"misw-api/db"
	"net/http"
)

type UserStatsResponse struct {
	Username     string  `json:"username"`
	GamesPlayed  int     `json:"gamesPlayed"`
	GamesWon     int     `json:"gamesWon"`
	GamesLost    int     `json:"gamesLost"`
	AverageMoves float64 `json:"averageMoves"`
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	conn := db.OpenConnection()
	defer db.CloseConnection(conn)
	user := auth.ValidateRequestingUser(w, r, conn)
	if user == nil {
		return
	}

	games, err := db.GetGamesForUser(conn, user.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve games", http.StatusInternalServerError)
		return
	}

	gamesPlayed := len(games)
	var gamesWon, gamesLost int
	var totalMoves int
	for _, game := range games {
		totalMoves += game.MovesCount
		if game.Won != nil {
			if *game.Won {
				gamesWon++
			} else {
				gamesLost++
			}
		}
	}

	averageMoves := 0.0
	if gamesPlayed > 0 {
		averageMoves = float64(totalMoves) / float64(gamesPlayed)
	}

	response := UserStatsResponse{
		Username:     user.Username,
		GamesPlayed:  gamesPlayed,
		GamesWon:     gamesWon,
		GamesLost:    gamesLost,
		AverageMoves: averageMoves,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}