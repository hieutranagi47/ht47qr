package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestShortURLRules(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s, err := NewShortURL(uuid.New(), uuid.New(), "https://example.com/a", "Ab012345", "request-1", now, 30*24*time.Hour)
	require.NoError(t, err)
	require.True(t, s.IsActiveAt(now.Add(29*24*time.Hour)))
	require.ErrorIs(t, func() error { _, err := s.ResolveAt(now.Add(30 * 24 * time.Hour)); return err }(), ErrInactive)
	require.Error(t, func() error {
		_, err := NewShortURL(uuid.New(), uuid.New(), "ftp://example.com", "Ab012345", "x", now, time.Hour)
		return err
	}())
}

func TestPlanLimits(t *testing.T) {
	limits, err := Pro.Limits()
	require.NoError(t, err)
	require.Equal(t, 100, limits.MaxLive)
	require.Error(t, func() error { _, err := Plan("x").Limits(); return err }())
}
