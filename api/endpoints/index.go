package endpoints

import (
	"encoding/json"
	"net/http"
)

func IndexHandlerFactory(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonBytes, err := json.Marshal(struct {
			Version string `json:"version"`
			Message string `json:"message"`
		}{
			Message: "Welcome to the MISW API!",
			Version: version,
		})
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	}
}