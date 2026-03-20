package main

import (
	"net/http"

	"github.com/omarraf/web-scraper/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

// middlewareAuth is a higher-order function that wraps handlers requiring authentication
// It extracts and validates the API key, looks up the user, and passes the user to the handler
func (apiCfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			respondWithError(w, 401, "Authorization header is required")
			return
		}

		if len(authHeader) < 8 || authHeader[:7] != "ApiKey " {
			respondWithError(w, 401, "Invalid authorization header format")
			return
		}
		apiKey := authHeader[7:] // This slices the string starting at position 7

		user, err := apiCfg.DB.GetUserByAPIKey(r.Context(), apiKey)
		if err != nil {
			// If we can't find the user, their API key is invalid
			respondWithError(w, 403, "Invalid API key")
			return
		}
		handler(w, r, user)
	}
}
