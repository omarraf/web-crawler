package main

import (
	"net/http"
	"strconv"

	"github.com/omarraf/web-scraper/internal/database"
)

func (apiCfg *apiConfig) handlerGetPosts(w http.ResponseWriter, r *http.Request, user database.User) {
	// Get limit query parameter (default to 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil {
			limit = limitInt
		}
	}

	// Get posts for this user from the database
	posts, err := apiCfg.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, 500, "Failed to get posts")
		return
	}

	// Convert database posts to response posts
	responsePosts := make([]Post, len(posts))
	for i, post := range posts {
		responsePosts[i] = databasePostToPost(post)
	}

	respondWithJSON(w, 200, responsePosts)
}
