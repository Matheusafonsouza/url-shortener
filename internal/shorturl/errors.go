package shorturl

import "errors"

var (
	ErrNotFound      = errors.New("short url not found")
	ErrCodeConflict  = errors.New("short code already exists")
	ErrExpired       = errors.New("short url expired")
	ErrInvalidURL    = errors.New("invalid url")
	ErrInvalidAlias  = errors.New("invalid alias")
	ErrReservedAlias = errors.New("alias is reserved")
	ErrInvalidTTL    = errors.New("invalid ttl")
)
