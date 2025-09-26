// main.go - HTTP server for DelayedNotifier

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

// Notification represents a delayed notification.
type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	Channel   string    `json:"channel"`
	SendAt    time.Time `json:"send_at"`
	Status    string    `json:"status"` // pending, sent, failed, cancelled
	Retries   int       `json:"retries"`
	CreatedAt time.Time `json:"created_at"`
}

// DelayedNotifier manages delayed notifications.
type DelayedNotifier struct {
	mu            sync.RWMutex
	notifications map[string]*Notification
	redis         *redis.Client
	rabbitConn    *amqp.Connection
	rabbitCh      *amqp.Channel
	rabbitQueue   amqp.Queue
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewDelayedNotifier creates a new DelayedNotifier instance.
func NewDelayedNotifier(redisAddr, rabbitAddr string) (*DelayedNotifier, error) {
	d := &DelayedNotifier{
		notifications: make(map[string]*Notification),
		redis: redis.NewClient(&redis.Options{
			Addr: redisAddr,
		}),
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	d.rabbitConn = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %v", err)
	}
	d.rabbitCh = ch

	q, err := ch.QueueDeclare(
		"notifications", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}
	d.rabbitQueue = q

	// Start worker
	d.ctx, d.cancel = context.WithCancel(context.Background())
	go d.worker()

	return d, nil
}

// Close closes connections.
func (d *DelayedNotifier) Close() {
	d.cancel()
	if d.rabbitCh != nil {
		d.rabbitCh.Close()
	}
	if d.rabbitConn != nil {
		d.rabbitConn.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
}

// CreateNotification creates a new delayed notification.
func (d *DelayedNotifier) CreateNotification(userID, message, channel string, sendAt time.Time) (string, error) {
	if sendAt.Before(time.Now()) {
		return "", fmt.Errorf("send_at must be in the future")
	}

	id := fmt.Sprintf("%d-%s", time.Now().Unix(), userID)
	notification := &Notification{
		ID:        id,
		UserID:    userID,
		Message:   message,
		Channel:   channel,
		SendAt:    sendAt,
		Status:    "pending",
		Retries:   0,
		CreatedAt: time.Now(),
	}

	d.mu.Lock()
	d.notifications[id] = notification
	d.mu.Unlock()

	// Cache in Redis
	err := d.redis.Set(context.Background(), id, notification.Status, 0).Err()
	if err != nil {
		log.Printf("error caching status: %v", err)
	}

	// Publish to RabbitMQ with delay
	body, _ := json.Marshal(notification)
	err = d.rabbitCh.Publish(
		"",                 // exchange
		d.rabbitQueue.Name, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers: amqp.Table{
				"x-delay": sendAt.Sub(time.Now()).Milliseconds(),
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to publish to queue: %v", err)
	}

	return id, nil
}

// GetNotificationStatus returns the status of a notification.
func (d *DelayedNotifier) GetNotificationStatus(id string) (string, error) {
	// Check cache first
	status, err := d.redis.Get(context.Background(), id).Result()
	if err == nil {
		return status, nil
	}

	d.mu.RLock()
	defer d.mu.RUnlock()
	if notification, ok := d.notifications[id]; ok {
		return notification.Status, nil
	}
	return "", fmt.Errorf("notification not found")
}

// CancelNotification cancels a pending notification.
func (d *DelayedNotifier) CancelNotification(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if notification, ok := d.notifications[id]; ok && notification.Status == "pending" {
		notification.Status = "cancelled"
		// Remove from Redis cache
		d.redis.Del(context.Background(), id)
		return nil
	}
	return fmt.Errorf("notification not found or not pending")
}

// worker processes the notification queue.
func (d *DelayedNotifier) worker() {
	defer d.rabbitCh.Close()
	defer d.rabbitConn.Close()

	msgs, err := d.rabbitCh.Consume(
		d.rabbitQueue.Name, // queue
		"",                 // consumer
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		log.Fatal("error consuming queue: ", err)
	}

	for {
		select {
		case <-d.ctx.Done():
			return
		case msg := <-msgs:
			var notification Notification
			if err := json.Unmarshal(msg.Body, &notification); err != nil {
				log.Printf("error unmarshaling notification: %v", err)
				msg.Nack(false, false)
				continue
			}

			// Check if cancelled
			if notification.Status == "cancelled" {
				msg.Ack(false)
				continue
			}

			// Send notification
			if err := d.sendNotification(&notification); err != nil {
				if notification.Retries < 3 {
					notification.Retries++
					delay := time.Duration(1<<notification.Retries) * time.Second
					notification.SendAt = time.Now().Add(delay)
					// Republish with new delay
					body, _ := json.Marshal(notification)
					msg.Nack(false, true)
					d.rabbitCh.Publish(
						"", d.rabbitQueue.Name, false, false,
						amqp.Publishing{
							ContentType: "application/json",
							Body:        body,
							Headers: amqp.Table{
								"x-delay": delay.Milliseconds(),
							},
						},
					)
				} else {
					notification.Status = "failed"
					log.Printf("failed to send notification %s after %d retries", notification.ID, notification.Retries)
					msg.Ack(false)
				}
				continue
			}

			notification.Status = "sent"
			msg.Ack(false)
		}
	}
}

// sendNotification sends the notification via the specified channel.
func (d *DelayedNotifier) sendNotification(notification *Notification) error {
	switch notification.Channel {
	case "email":
		return d.sendEmail(notification)
	case "telegram":
		return d.sendTelegram(notification)
	default:
		return fmt.Errorf("unsupported channel: %s", notification.Channel)
	}
}

// sendEmail sends notification via email (mock implementation).
func (d *DelayedNotifier) sendEmail(notification *Notification) error {
	log.Printf("Sending email to user %s: %s", notification.UserID, notification.Message)
	return nil
}

// sendTelegram sends notification via Telegram (mock implementation).
func (d *DelayedNotifier) sendTelegram(notification *Notification) error {
	log.Printf("Sending Telegram message to user %s: %s", notification.UserID, notification.Message)
	return nil
}

// CreateNotificationHandler handles POST /notify.
func (d *DelayedNotifier) CreateNotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, `{"error": "bad request"}`, http.StatusBadRequest)
		return
	}

	userID := r.Form.Get("user_id")
	message := r.Form.Get("message")
	channel := r.Form.Get("channel")
	dateStr := r.Form.Get("send_at")

	date, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		http.Error(w, `{"error": "invalid send_at format, use YYYY-MM-DD HH:MM:SS"}`, http.StatusBadRequest)
		return
	}

	id, err := d.CreateNotification(userID, message, channel, date)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": id})
}

// GetNotificationHandler handles GET /notify/{id}.
func (d *DelayedNotifier) GetNotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/notify/")
	status, err := d.GetNotificationStatus(id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": status})
}

// CancelNotificationHandler handles DELETE /notify/{id}.
func (d *DelayedNotifier) CancelNotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/notify/")
	if err := d.CancelNotification(id); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": "cancelled"})
}

func main() {
	redisAddr := flag.String("redis", "localhost:6379", "Redis address")
	rabbitAddr := flag.String("rabbit", "amqp://guest:guest@localhost:5672/", "RabbitMQ address")
	flag.Parse()

	notifier, err := NewDelayedNotifier(*redisAddr, *rabbitAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer notifier.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/notify", notifier.CreateNotificationHandler)
	mux.HandleFunc("/notify/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			notifier.GetNotificationHandler(w, r)
		} else if r.Method == http.MethodDelete {
			notifier.CancelNotificationHandler(w, r)
		} else {
			http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	handler := LogMiddleware(mux)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// LogMiddleware logs HTTP requests.
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
