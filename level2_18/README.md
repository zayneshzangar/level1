# Calendar HTTP Server

A simple HTTP server for managing calendar events, implemented in Go. Supports CRUD operations for events, with endpoints for creating, updating, deleting, and retrieving events by day, week, or month.

## Features

- **CRUD Operations**:
  - `POST /create_event`: Create a new event (user_id, date, title).
  - `POST /update_event`: Update an existing event by ID.
  - `POST /delete_event`: Delete an event by ID.
  - `GET /events_for_day`: Get events for a specific day.
  - `GET /events_for_week`: Get events for a week starting from a date.
  - `GET /events_for_month`: Get events for a month starting from a date.
- **Request Format**:
  - POST requests use `application/x-www-form-urlencoded`.
  - GET requests use query string parameters.
- **Response Format**:
  - Success: `{"result": {...}}` with HTTP 200.
  - Input errors: `{"error": "..."}` with HTTP 400.
  - Business logic errors: `{"error": "..."}` with HTTP 503.
- **Storage**: In-memory (slice with mutex for thread-safety).
- **Middleware**: Logs method, URL, and request duration to stdout.
- **Configuration**: Port via `PORT` environment variable (default: 8080).
- **Tests**: Unit tests for business logic (add, update, delete, retrieve events).

## Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd calendar-server


Directory structure:
calendar-server/
├── main.go
├── calendar/
│   ├── calendar.go
│   └── calendar_test.go
└── README.md


Run the server:
PORT=8080 go run main.go


Run tests:
go test -v ./calendar


Verify code quality:
go vet ./...
golint ./...



Example Requests

Create Event:
curl -d "user_id=1&date=2025-09-20&title=Team Meeting" http://localhost:8080/create_event

Response:
{"result":{"id":0,"user_id":1,"date":"2025-09-20T00:00:00Z","title":"Team Meeting"}}


Update Event:
curl -d "id=0&date=2025-09-21&title=Updated Meeting" http://localhost:8080/update_event

Response:
{"result":"event updated"}


Delete Event:
curl -d "id=0" http://localhost:8080/delete_event

Response:
{"result":"event deleted"}


Get Events for Day:
curl "http://localhost:8080/events_for_day?user_id=1&date=2025-09-20"

Response:
{"result":[{"id":0,"user_id":1,"date":"2025-09-20T00:00:00Z","title":"Team Meeting"}]}


Get Events for Week:
curl "http://localhost:8080/events_for_week?user_id=1&date=2025-09-20"


Get Events for Month:
curl "http://localhost:8080/events_for_month?user_id=1&date=2025-09-01"



Notes

Thread Safety: Uses sync.Mutex to prevent data races when accessing the in-memory event slice.
Error Handling: Validates input parameters (user_id, date, title) and returns appropriate HTTP status codes.
Extensibility: Business logic is separated from HTTP handlers, making it easy to swap in-memory storage for a database.
JSON Support: To add JSON request support, modify handlers to check Content-Type and use json.NewDecoder.

Screenshots/Video

Screenshots: Include in screenshots/ directory (e.g., curl command outputs).
Video: Record a demo showing all endpoints (create, update, delete, get events) and save as demo.mp4.


