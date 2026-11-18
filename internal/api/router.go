package api

import (
	"encoding/json"
	"net/http"

	"github.com/maze-n/jsongo/internal/database"
)

func handleGETs(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{Status: "ok"})
}

func HandleRoutes(store *database.Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{path...}", JsonContentTypeMiddleware(handleGETs))

	return mux
}
