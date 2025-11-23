package auth

import (
	"encoding/base64"
	"misw-api/db"
	"misw-api/model"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

func ValidateRequestingUser(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) *model.User {
	header := r.Header.Get("Authorization")
	if header == "" || len(header) < 6 || header[:6] != "Basic " {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}

	payload, err := base64.StdEncoding.DecodeString(header[6:])
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}

	username := pair[0]
	password := pair[1]

	user, err := db.GetUser(conn, username, password)
	if(err != nil){
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil;
	}

	return user;
}