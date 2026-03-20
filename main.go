package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/omarraf/web-scraper/internal/database"
)

type apiConfig struct {
	DB           *database.Queries
	CrawlEngines sync.Map // jobID (uuid.UUID) -> context.CancelFunc
}

func main() {

	godotenv.Load(".env")
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}
	fmt.Println("PORT:", portString)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not found in the environment")
	}
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot open database:", err)
	}
	// Neon's pgBouncer pooler (transaction mode) doesn't support prepared
	// statements across connections. Pin to a single connection to avoid
	// "bind message has N result formats but query has M columns" errors.
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	apiCfg := apiConfig{
		DB: database.New(conn),
	}

	db, err := initDB(dbURL)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer db.Close()

	// Start the scraper in a goroutine
	go startScraping(apiCfg.DB, 10, time.Minute)

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
	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))

	// Feed routes
	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)

	// Feed follow routes
	v1Router.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))

	// Posts route
	v1Router.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetPosts))

	// Crawl job routes
	v1Router.Post("/crawl_jobs", apiCfg.middlewareAuth(apiCfg.handlerCreateCrawlJob))
	v1Router.Get("/crawl_jobs", apiCfg.middlewareAuth(apiCfg.handlerListCrawlJobs))
	v1Router.Get("/crawl_jobs/{jobID}", apiCfg.middlewareAuth(apiCfg.handlerGetCrawlJob))
	v1Router.Delete("/crawl_jobs/{jobID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteCrawlJob))

	// Graph + analysis routes
	v1Router.Get("/crawl_jobs/{jobID}/graph", apiCfg.middlewareAuth(apiCfg.handlerGetGraph))
	v1Router.Get("/crawl_jobs/{jobID}/analysis", apiCfg.middlewareAuth(apiCfg.handlerGetAnalysis))
	v1Router.Get("/crawl_jobs/{jobID}/pagerank", apiCfg.middlewareAuth(apiCfg.handlerGetPageRank))

	router.Mount("/v1", v1Router)

	// Serve the D3 frontend at /
	router.Handle("/*", http.FileServer(http.Dir("web")))

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
