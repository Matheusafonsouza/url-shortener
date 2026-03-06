package shorturl

import "time"

type URLMapping struct {
	Code           string
	OriginalURL    string
	CreatedAt      time.Time
	ExpiresAt      *time.Time
	ClickCount     int64
	LastAccessedAt *time.Time
}

type CreateInput struct {
	OriginalURL string
	Alias       string
	TTLHours    *int
}
