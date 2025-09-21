package calendar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Event represents a calendar event.
type Event struct {
	ID     int       `json:"id"`
	UserID int       `json:"user_id"`
	Date   time.Time `json:"date"`
	Title  string    `json:"title"`
}

// Calendar manages events in memory.
type Calendar struct {
	mu     sync.Mutex
	events []*Event
	nextID int
}

// NewCalendar creates a new Calendar instance.
func NewCalendar() *Calendar {
	return &Calendar{}
}

// AddEvent adds a new event to the calendar.
func (c *Calendar) AddEvent(userID int, date time.Time, title string) *Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	event := &Event{
		ID:     c.nextID,
		UserID: userID,
		// Normalize to UTC and truncate to date
		Date:  date.Truncate(24 * time.Hour).UTC(),
		Title: title,
	}
	c.nextID++
	c.events = append(c.events, event)
	return event
}

// UpdateEvent updates an existing event by ID.
func (c *Calendar) UpdateEvent(id int, date time.Time, title string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, event := range c.events {
		if event.ID == id {
			// Normalize to UTC and truncate to date
			event.Date = date.Truncate(24 * time.Hour).UTC()
			event.Title = title
			return nil
		}
	}
	return fmt.Errorf("event with ID %d not found", id)
}

// DeleteEvent deletes an event by ID.
func (c *Calendar) DeleteEvent(id int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, event := range c.events {
		if event.ID == id {
			c.events = append(c.events[:i], c.events[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("event with ID %d not found", id)
}

// GetEventsForDay returns events for a specific user and day.
func (c *Calendar) GetEventsForDay(userID int, date time.Time) []*Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	var result []*Event
	// Normalize input date
	targetDate := date.Truncate(24 * time.Hour).UTC()
	for _, event := range c.events {
		eventDate := event.Date.Truncate(24 * time.Hour).UTC()
		if event.UserID == userID &&
			eventDate.Year() == targetDate.Year() &&
			eventDate.Month() == targetDate.Month() &&
			eventDate.Day() == targetDate.Day() {
			result = append(result, event)
		}
	}
	return result
}

// GetEventsForWeek returns events for a specific user and week (starting from date).
func (c *Calendar) GetEventsForWeek(userID int, date time.Time) []*Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	var result []*Event
	// Normalize input date
	start := date.Truncate(24 * time.Hour).UTC()
	end := start.Add(7 * 24 * time.Hour)
	for _, event := range c.events {
		eventDate := event.Date.Truncate(24 * time.Hour).UTC()
		if event.UserID == userID && !eventDate.Before(start) && eventDate.Before(end) {
			result = append(result, event)
		}
	}
	return result
}

// GetEventsForMonth returns events for a specific user and month (starting from date).
func (c *Calendar) GetEventsForMonth(userID int, date time.Time) []*Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	var result []*Event
	// Normalize input date
	start := date.Truncate(24 * time.Hour).UTC()
	end := start.AddDate(0, 1, 0)
	for _, event := range c.events {
		eventDate := event.Date.Truncate(24 * time.Hour).UTC()
		if event.UserID == userID && !eventDate.Before(start) && eventDate.Before(end) {
			result = append(result, event)
		}
	}
	return result
}

// CreateEventHandler handles POST /create_event.
func (c *Calendar) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, `{"error": "bad request"}`, http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.Form.Get("user_id"))
	if err != nil {
		http.Error(w, `{"error": "invalid user_id"}`, http.StatusBadRequest)
		return
	}

	dateStr := r.Form.Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, `{"error": "invalid date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	title := r.Form.Get("title")
	if title == "" {
		http.Error(w, `{"error": "title is required"}`, http.StatusBadRequest)
		return
	}

	event := c.AddEvent(userID, date, title)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]*Event{"result": event})
}

// UpdateEventHandler handles POST /update_event.
func (c *Calendar) UpdateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, `{"error": "bad request"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}

	dateStr := r.Form.Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, `{"error": "invalid date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	title := r.Form.Get("title")
	if title == "" {
		http.Error(w, `{"error": "title is required"}`, http.StatusBadRequest)
		return
	}

	if err := c.UpdateEvent(id, date, title); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": "event updated"})
}

// DeleteEventHandler handles POST /delete_event.
func (c *Calendar) DeleteEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, `{"error": "bad request"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}

	if err := c.DeleteEvent(id); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": "event deleted"})
}

// EventsForDayHandler handles GET /events_for_day.
func (c *Calendar) EventsForDayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil {
		http.Error(w, `{"error": "invalid user_id"}`, http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, `{"error": "invalid date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	events := c.GetEventsForDay(userID, date)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]*Event{"result": events})
}

// EventsForWeekHandler handles GET /events_for_week.
func (c *Calendar) EventsForWeekHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil {
		http.Error(w, `{"error": "invalid user_id"}`, http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, `{"error": "invalid date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	events := c.GetEventsForWeek(userID, date)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]*Event{"result": events})
}

// EventsForMonthHandler handles GET /events_for_month.
func (c *Calendar) EventsForMonthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil {
		http.Error(w, `{"error": "invalid user_id"}`, http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, `{"error": "invalid date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	events := c.GetEventsForMonth(userID, date)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]*Event{"result": events})
}
