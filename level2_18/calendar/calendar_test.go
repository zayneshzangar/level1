package calendar

import (
	"testing"
	"time"
)

func TestNewCalendar(t *testing.T) {
	c := NewCalendar()
	if c == nil {
		t.Error("NewCalendar returned nil")
	}
	if len(c.events) != 0 {
		t.Errorf("Expected empty events, got %d", len(c.events))
	}
}

func TestAddEvent(t *testing.T) {
	c := NewCalendar()
	date := time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC)
	event := c.AddEvent(1, date, "Test Meeting")
	if event.ID != 0 {
		t.Errorf("Expected ID 0, got %d", event.ID)
	}
	if event.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", event.UserID)
	}
	if event.Title != "Test Meeting" {
		t.Errorf("Expected Title 'Test Meeting', got %s", event.Title)
	}
	if !event.Date.Equal(date) {
		t.Errorf("Expected Date %v, got %v", date, event.Date)
	}
	if len(c.events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(c.events))
	}
}

func TestUpdateEvent(t *testing.T) {
	c := NewCalendar()
	date := time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC)
	event := c.AddEvent(1, date, "Old Meeting")
	newDate := date.Add(24 * time.Hour)
	if err := c.UpdateEvent(event.ID, newDate, "New Meeting"); err != nil {
		t.Errorf("UpdateEvent failed: %v", err)
	}
	updatedEvent := c.events[0]
	if updatedEvent.Title != "New Meeting" {
		t.Errorf("Expected Title 'New Meeting', got %s", updatedEvent.Title)
	}
	if !updatedEvent.Date.Equal(newDate) {
		t.Errorf("Expected Date %v, got %v", newDate, updatedEvent.Date)
	}
	if err := c.UpdateEvent(999, newDate, "Invalid"); err == nil {
		t.Error("UpdateEvent should fail for non-existent event")
	}
}

func TestDeleteEvent(t *testing.T) {
	c := NewCalendar()
	date := time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC)
	event := c.AddEvent(1, date, "Test Meeting")
	if err := c.DeleteEvent(event.ID); err != nil {
		t.Errorf("DeleteEvent failed: %v", err)
	}
	if len(c.events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(c.events))
	}
	if err := c.DeleteEvent(event.ID); err == nil {
		t.Error("DeleteEvent should fail for non-existent event")
	}
}

func TestGetEventsForDay(t *testing.T) {
	c := NewCalendar()
	date := time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC)
	c.AddEvent(1, date, "Event 1")
	c.AddEvent(1, date.Add(24*time.Hour), "Event 2")
	c.AddEvent(2, date, "Event 3")
	events := c.GetEventsForDay(1, date)
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if len(events) > 0 && events[0].Title != "Event 1" {
		t.Errorf("Expected Title 'Event 1', got %s", events[0].Title)
	}
}

func TestGetEventsForWeek(t *testing.T) {
	c := NewCalendar()
	date := time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC)
	c.AddEvent(1, date, "Event 1")
	c.AddEvent(1, date.Add(3*24*time.Hour), "Event 2")
	c.AddEvent(1, date.Add(8*24*time.Hour), "Event 3")
	events := c.GetEventsForWeek(1, date)
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

func TestGetEventsForMonth(t *testing.T) {
	c := NewCalendar()
	date := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	c.AddEvent(1, date, "Event 1")
	c.AddEvent(1, date.Add(15*24*time.Hour), "Event 2")
	c.AddEvent(1, date.Add(32*24*time.Hour), "Event 3")
	events := c.GetEventsForMonth(1, date)
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}
