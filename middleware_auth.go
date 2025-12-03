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
		// Step 1: Extract the Authorization header from the HTTP request
		// Headers are key-value pairs sent with every HTTP request
		// Example: "Authorization: ApiKey abc123xyz"
		authHeader := r.Header.Get("Authorization")

		// Step 2: Check if the header is empty
		// If no header was sent, the user isn't trying to authenticate
		if authHeader == "" {
			respondWithError(w, 401, "Authorization header is required")
			return
		}

		// Step 3: Validate the format - it should be "ApiKey <key>"
		// len("ApiKey ") = 7, so we need at least 8 characters (space + 1 char key)
		if len(authHeader) < 8 || authHeader[:7] != "ApiKey " {
			respondWithError(w, 401, "Invalid authorization header format")
			return
		}

		// Step 4: Extract just the API key part (everything after "ApiKey ")
		// Example: "ApiKey abc123" becomes "abc123"
		apiKey := authHeader[7:] // This slices the string starting at position 7

		// Step 5: Look up the user in our database using this API key
		// r.Context() carries request-specific values like deadlines/cancellation
		user, err := apiCfg.DB.GetUserByAPIKey(r.Context(), apiKey)
		if err != nil {
			// If we can't find the user, their API key is invalid
			respondWithError(w, 403, "Invalid API key")
			return
		}

		// Step 6: Success! Call the actual handler and pass the authenticated user
		// This is why we return http.HandlerFunc - we're creating a new handler
		// that does auth checks before calling the real handler
		handler(w, r, user)
	}
}
