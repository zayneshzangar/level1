package shortener

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// Shortener manages URL shortening and analytics.
type Shortener struct {
	db    *sql.DB
	redis *redis.Client
	mu    sync.Mutex
}

// Click represents a single click on a short URL.
type Click struct {
	Timestamp time.Time `json:"timestamp"`
	UserAgent string    `json:"user_agent"`
	ShortURL  string    `json:"short_url"`
}

// Analytics represents aggregated click data.
type Analytics struct {
	TotalClicks int            `json:"total_clicks"`
	ByDay       map[string]int `json:"by_day"`
	ByMonth     map[string]int `json:"by_month"`
	ByUserAgent map[string]int `json:"by_user_agent"`
}

// NewShortener creates a new Shortener instance.
func NewShortener(db *sql.DB, redis *redis.Client) (*Shortener, error) {
	s := &Shortener{db: db, redis: redis}
	if err := s.initDB(); err != nil {
		return nil, err
	}
	return s, nil
}

// initDB creates necessary tables.
func (s *Shortener) initDB() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			short_url VARCHAR(50) PRIMARY KEY,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
		CREATE TABLE IF NOT EXISTS clicks (
			id SERIAL PRIMARY KEY,
			short_url VARCHAR(50) REFERENCES urls(short_url),
			timestamp TIMESTAMP NOT NULL,
			user_agent TEXT
		);
	`)
	return err
}

// Shorten creates a new short URL.
func (s *Shortener) Shorten(originalURL, customShort string) (string, error) {
	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		return "", fmt.Errorf("invalid URL: must start with http:// or https://")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate short URL if not custom
	shortURL := customShort
	if shortURL == "" {
		b := make([]byte, 6)
		_, err := rand.Read(b)
		if err != nil {
			return "", fmt.Errorf("failed to generate short URL: %v", err)
		}
		shortURL = base64.RawURLEncoding.EncodeToString(b)
	}

	// Check if short URL exists
	var exists string
	err := s.db.QueryRow("SELECT short_url FROM urls WHERE short_url = $1", shortURL).Scan(&exists)
	if err == nil {
		return "", fmt.Errorf("short URL already exists")
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("database error: %v", err)
	}

	// Insert into database
	_, err = s.db.Exec("INSERT INTO urls (short_url, original_url, created_at) VALUES ($1, $2, $3)",
		shortURL, originalURL, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to save URL: %v", err)
	}

	// Cache in Redis
	err = s.redis.Set(context.Background(), shortURL, originalURL, 24*time.Hour).Err()
	if err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to cache URL in Redis: %v\n", err)
	}

	return shortURL, nil
}

// GetOriginalURL retrieves the original URL and logs a click.
func (s *Shortener) GetOriginalURL(shortURL string, userAgent string) (string, error) {
	// Check Redis cache
	originalURL, err := s.redis.Get(context.Background(), shortURL).Result()
	if err == nil {
		// Log click
		go s.logClick(shortURL, userAgent)
		return originalURL, nil
	}

	// Fallback to database
	var url string
	err = s.db.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortURL).Scan(&url)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("short URL not found")
	}
	if err != nil {
		return "", fmt.Errorf("database error: %v", err)
	}

	// Cache in Redis
	err = s.redis.Set(context.Background(), shortURL, url, 24*time.Hour).Err()
	if err != nil {
		fmt.Printf("Warning: failed to cache URL in Redis: %v\n", err)
	}

	// Log click
	go s.logClick(shortURL, userAgent)
	return url, nil
}

// logClick saves click data to the database.
func (s *Shortener) logClick(shortURL, userAgent string) {
	_, err := s.db.Exec("INSERT INTO clicks (short_url, timestamp, user_agent) VALUES ($1, $2, $3)",
		shortURL, time.Now(), userAgent)
	if err != nil {
		fmt.Printf("Warning: failed to log click: %v\n", err)
	}
}

// GetAnalytics retrieves analytics for a short URL.
func (s *Shortener) GetAnalytics(shortURL string) (*Analytics, error) {
	// Check if short URL exists
	var exists string
	err := s.db.QueryRow("SELECT short_url FROM urls WHERE short_url = $1", shortURL).Scan(&exists)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("short URL not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Query clicks
	rows, err := s.db.Query("SELECT timestamp, user_agent FROM clicks WHERE short_url = $1", shortURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query clicks: %v", err)
	}
	defer rows.Close()

	analytics := &Analytics{
		ByDay:       make(map[string]int),
		ByMonth:     make(map[string]int),
		ByUserAgent: make(map[string]int),
	}

	for rows.Next() {
		var click Click
		if err := rows.Scan(&click.Timestamp, &click.UserAgent); err != nil {
			return nil, fmt.Errorf("failed to scan click: %v", err)
		}
		analytics.TotalClicks++
		analytics.ByDay[click.Timestamp.Format("2006-01-02")]++
		analytics.ByMonth[click.Timestamp.Format("2006-01")]++
		analytics.ByUserAgent[click.UserAgent]++
	}

	return analytics, nil
}

// ShortenHandler handles POST /shorten.
func (s *Shortener) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, `{"error": "bad request"}`, http.StatusBadRequest)
		return
	}

	originalURL := r.Form.Get("original_url")
	customShort := r.Form.Get("custom_short")

	shortURL, err := s.Shorten(originalURL, customShort)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": shortURL})
}

// RedirectHandler handles GET /s/{short_url}.
func (s *Shortener) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	shortURL := strings.TrimPrefix(r.URL.Path, "/s/")
	if shortURL == "" {
		http.Error(w, `{"error": "short URL required"}`, http.StatusBadRequest)
		return
	}

	originalURL, err := s.GetOriginalURL(shortURL, r.UserAgent())
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// AnalyticsHandler handles GET /analytics/{short_url}.
func (s *Shortener) AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	shortURL := strings.TrimPrefix(r.URL.Path, "/analytics/")
	if shortURL == "" {
		http.Error(w, `{"error": "short URL required"}`, http.StatusBadRequest)
		return
	}

	analytics, err := s.GetAnalytics(shortURL)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]*Analytics{"result": analytics})
}

// UIHandler serves the HTML UI.
func (s *Shortener) UIHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "static/index.html")
}
