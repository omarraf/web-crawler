package main

import (
	"net/http"

	"github.com/omarraf/web-scraper/internal/database"
)

// handlerGetUser returns the authenticated user's information
// Notice it takes 3 parameters: w, r, AND user
// The middleware already looked up the user, so we just return it!
func (apiCfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	// Convert the database user model to our API response format
	// databaseUserToUser() is a helper function defined in models.go
	// It converts from the internal database type to what we send to clients
	respondWithJSON(w, 200, databaseUserToUser(user))
}
