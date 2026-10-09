package domain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidLongURL = errors.New("long URL must be an absolute HTTP or HTTPS URL")
	ErrInvalidCode    = errors.New("short code must contain exactly 8 base62 characters")
	ErrInactive       = errors.New("short URL is inactive")
)

type ShortURL struct {
	id             uuid.UUID
	userID         uuid.UUID
	longURL        string
	code           string
	idempotencyKey string
	createdAt      time.Time
	expiresAt      time.Time
	deactivatedAt  *time.Time
}

func NewShortURL(id, userID uuid.UUID, longURL, code, idempotencyKey string, now time.Time, lifetime time.Duration) (ShortURL, error) {
	if id == uuid.Nil {
		return ShortURL{}, errors.New("id is required")
	}
	if !validLongURL(longURL) {
		return ShortURL{}, ErrInvalidLongURL
	}
	if !validCode(code) {
		return ShortURL{}, ErrInvalidCode
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		return ShortURL{}, errors.New("idempotency key is required")
	}
	if lifetime <= 0 {
		return ShortURL{}, errors.New("URL lifetime must be positive")
	}
	now = now.UTC()
	return ShortURL{id: id, userID: userID, longURL: longURL, code: code, idempotencyKey: idempotencyKey, createdAt: now, expiresAt: now.Add(lifetime)}, nil
}

func RestoreShortURL(id, userID uuid.UUID, longURL, code, idempotencyKey string, createdAt, expiresAt time.Time, deactivatedAt *time.Time) (ShortURL, error) {
	if id == uuid.Nil || !validLongURL(longURL) || !validCode(code) || strings.TrimSpace(idempotencyKey) == "" || expiresAt.Before(createdAt) {
		return ShortURL{}, errors.New("invalid short URL state")
	}
	return ShortURL{id: id, userID: userID, longURL: longURL, code: code, idempotencyKey: idempotencyKey, createdAt: createdAt.UTC(), expiresAt: expiresAt.UTC(), deactivatedAt: copyTime(deactivatedAt)}, nil
}

func (s ShortURL) ID() uuid.UUID             { return s.id }
func (s ShortURL) UserID() uuid.UUID         { return s.userID }
func (s ShortURL) LongURL() string           { return s.longURL }
func (s ShortURL) Code() string              { return s.code }
func (s ShortURL) IdempotencyKey() string    { return s.idempotencyKey }
func (s ShortURL) CreatedAt() time.Time      { return s.createdAt }
func (s ShortURL) ExpiresAt() time.Time      { return s.expiresAt }
func (s ShortURL) DeactivatedAt() *time.Time { return copyTime(s.deactivatedAt) }

func (s ShortURL) IsActiveAt(now time.Time) bool {
	return s.deactivatedAt == nil && now.Before(s.expiresAt)
}

func (s ShortURL) ResolveAt(now time.Time) (string, error) {
	if !s.IsActiveAt(now) {
		return "", ErrInactive
	}
	return s.longURL, nil
}

func (s *ShortURL) Deactivate(now time.Time) error {
	if s.deactivatedAt != nil {
		return ErrInactive
	}
	now = now.UTC()
	s.deactivatedAt = &now
	return nil
}

func validLongURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}

func validCode(code string) bool {
	if len(code) != 8 {
		return false
	}
	for _, r := range code {
		if !(r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') {
			return false
		}
	}
	return true
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	v := value.UTC()
	return &v
}

func (s ShortURL) String() string { return fmt.Sprintf("%s/%s", s.id, s.code) }
