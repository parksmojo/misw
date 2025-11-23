package endpoints

import (
	"encoding/json"
	"misw-api/db"
	"net/http"
	"strings"
)

type requestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn := db.OpenConnection()
	defer db.CloseConnection(conn)
	
	_, err := db.CreateUser(conn, body.Username, body.Password);
	if err != nil {
		if strings.Contains(err.Error(), "username") {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Unknown error occured", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}