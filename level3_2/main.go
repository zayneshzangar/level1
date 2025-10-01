package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"level32/shortener"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

// main starts the URL shortener server.
func main() {
	// Get configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		log.Fatal("POSTGRES_DSN not set")
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer db.Close()

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer redisClient.Close()

	// Initialize shortener
	s, err := shortener.NewShortener(db, redisClient)
	if err != nil {
		log.Fatal("Failed to initialize shortener:", err)
	}

	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", s.ShortenHandler)
	mux.HandleFunc("/s/", s.RedirectHandler)
	mux.HandleFunc("/analytics/", s.AnalyticsHandler)
	mux.HandleFunc("/", s.UIHandler)

	// Apply logging middleware
	handler := LogMiddleware(mux)

	// Start server
	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// LogMiddleware logs HTTP requests.
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
