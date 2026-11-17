package api

import "net/http"

func HandleRoutes() http.Handler {
	mux := http.NewServeMux()

	return mux

}
