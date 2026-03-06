package shorturl

import (
	"context"
	"crypto/rand"
	"math/big"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	generatedCodeLength = 8
	maxCodeRetries      = 5
)

var aliasPattern = regexp.MustCompile(`^[a-z0-9_-]{4,32}$`)

var reservedAliases = map[string]struct{}{
	"api":    {},
	"health": {},
	"shorten": {},
	"links":  {},
}

type Service struct {
	repo       Repository
	defaultTTL time.Duration
	nowFunc    func() time.Time
}

func NewService(repo Repository, defaultTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		defaultTTL: defaultTTL,
		nowFunc:    time.Now,
	}
}

func (s *Service) CreateShortURL(ctx context.Context, input CreateInput) (URLMapping, error) {
	originalURL := strings.TrimSpace(input.OriginalURL)
	if !isValidURL(originalURL) {
		return URLMapping{}, ErrInvalidURL
	}

	alias := normalizeCode(input.Alias)
	if alias != "" {
		if !aliasPattern.MatchString(alias) {
			return URLMapping{}, ErrInvalidAlias
		}

		if isReservedAlias(alias) {
			return URLMapping{}, ErrReservedAlias
		}
	}

	now := s.nowFunc().UTC()
	expiresAt, err := s.computeExpiresAt(now, input.TTLHours)
	if err != nil {
		return URLMapping{}, err
	}

	if alias != "" {
		mapping := URLMapping{
			Code:        alias,
			OriginalURL: originalURL,
			CreatedAt:   now,
			ExpiresAt:   expiresAt,
		}

		err := s.repo.Create(ctx, mapping)
		if err != nil {
			return URLMapping{}, err
		}

		return mapping, nil
	}

	for attempt := 0; attempt < maxCodeRetries; attempt++ {
		code, err := generateCode(generatedCodeLength)
		if err != nil {
			return URLMapping{}, err
		}

		if isReservedAlias(code) {
			continue
		}

		mapping := URLMapping{
			Code:        code,
			OriginalURL: originalURL,
			CreatedAt:   now,
			ExpiresAt:   expiresAt,
		}

		err = s.repo.Create(ctx, mapping)
		if err == ErrCodeConflict {
			continue
		}

		if err != nil {
			return URLMapping{}, err
		}

		return mapping, nil
	}

	return URLMapping{}, ErrCodeConflict
}

func (s *Service) ResolveShortCode(ctx context.Context, code string) (URLMapping, error) {
	normalizedCode := normalizeCode(code)
	mapping, err := s.repo.GetByCode(ctx, normalizedCode)
	if err != nil {
		return URLMapping{}, err
	}

	if mapping.ExpiresAt != nil && s.nowFunc().UTC().After(mapping.ExpiresAt.UTC()) {
		return URLMapping{}, ErrExpired
	}

	if err := s.repo.IncrementClickCount(ctx, normalizedCode); err != nil {
		return URLMapping{}, err
	}

	mapping.ClickCount++
	now := s.nowFunc().UTC()
	mapping.LastAccessedAt = &now

	return mapping, nil
}

func (s *Service) GetStats(ctx context.Context, code string) (URLMapping, error) {
	normalizedCode := normalizeCode(code)
	mapping, err := s.repo.GetByCode(ctx, normalizedCode)
	if err != nil {
		return URLMapping{}, err
	}

	if mapping.ExpiresAt != nil && s.nowFunc().UTC().After(mapping.ExpiresAt.UTC()) {
		return URLMapping{}, ErrExpired
	}

	return mapping, nil
}

func (s *Service) computeExpiresAt(now time.Time, ttlHours *int) (*time.Time, error) {
	ttl := s.defaultTTL
	if ttlHours != nil {
		if *ttlHours <= 0 {
			return nil, ErrInvalidTTL
		}

		ttl = time.Duration(*ttlHours) * time.Hour
	}

	expiresAt := now.Add(ttl)
	return &expiresAt, nil
}

func normalizeCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}

func isReservedAlias(alias string) bool {
	_, exists := reservedAliases[alias]
	return exists
}

func isValidURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}

	return true
}

func generateCode(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	builder := strings.Builder{}
	builder.Grow(length)

	for idx := 0; idx < length; idx++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		builder.WriteByte(charset[randomIndex.Int64()])
	}

	return builder.String(), nil
}
