package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load(".env")
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found ni the environment")
	}
	fmt.Println("PORT:", portString)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not found in the environment")
	}

	db, err := initDB(dbURL)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer db.Close()

	router := chi.NewRouter()

	// send extra http headers
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
	v1Router.HandleFunc("/healthz", handlerReadiness)
	v1Router.Get("/err", handleErr)

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	log.Printf("Server is starting on port %s\n", portString)

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
