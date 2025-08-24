package main

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
)

// getNTPTime запрашивает текущее время через NTP-сервер.
func getNTPTime() (time.Time, error) {
	resp, err := ntp.Query("2.kz.pool.ntp.org")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to query NTP server: %w", err)
	}
	// Корректируем время с учётом задержкиd
	return resp.Time, nil
}

func main() {
	// Получаем время через NTP
	ntpTime, err := getNTPTime()
	if err != nil {
		// Выводим ошибку в stderr и завершаем с ненулевым кодом
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Выводим текущее время
	fmt.Printf("Current time (NTP): %s\n", ntpTime.Format(time.RFC3339))
}
