package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"level218/calendar"
)

// main starts the HTTP server.
func main() {
	// Get port from environment variable or default to 8080
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("Invalid port")
	}

	// Initialize calendar
	cal := calendar.NewCalendar()

	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/create_event", cal.CreateEventHandler)
	mux.HandleFunc("/update_event", cal.UpdateEventHandler)
	mux.HandleFunc("/delete_event", cal.DeleteEventHandler)
	mux.HandleFunc("/events_for_day", cal.EventsForDayHandler)
	mux.HandleFunc("/events_for_week", cal.EventsForWeekHandler)
	mux.HandleFunc("/events_for_month", cal.EventsForMonthHandler)

	// Apply logging middleware
	handler := LogMiddleware(mux)

	// Start server
	log.Printf("Server starting on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}

// LogMiddleware logs HTTP requests (method, URL, duration).
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
