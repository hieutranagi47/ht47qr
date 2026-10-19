package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRoomValidationAndLifetime(t *testing.T) {
	now := time.Now().UTC()
	token := "private"
	host, err := RestorePlayer("host", " Alice ", "", Host, HashCredential(token), now, "192.0.2.1")
	require.NoError(t, err)
	require.Equal(t, "Alice", host.Name())
	for _, n := range []int{1, 3, 5, 7} {
		r, err := NewRoom("room", " abc234 ", "PUBLIC", n, host, now, "192.0.2.1")
		require.NoError(t, err)
		require.Equal(t, "ABC234", r.Code())
		require.True(t, r.Active(now.Add(Lifetime-time.Nanosecond)))
		require.False(t, r.Active(now.Add(Lifetime)))
	}
	for _, n := range []int{0, 2, 4, 6, 8} {
		_, err := NewRoom("r", "ABC234", "PUBLIC", n, host, now, "192.0.2.1")
		require.ErrorIs(t, err, ErrInvalid)
	}
	for _, name := range []string{"", "  ", strings.Repeat("界", 51), string([]byte{0xff})} {
		_, err := RestorePlayer("id", name, "", Guest, "", now, "192.0.2.1")
		require.ErrorIs(t, err, ErrInvalid)
	}
	_, err = RestorePlayer("id", strings.Repeat("界", 50), "", Guest, "", now, "192.0.2.1")
	require.NoError(t, err)
	for _, code := range []string{"short", "ABC123", "ABC23!", "ABC2345"} {
		_, err := NormalizeCode(code)
		require.ErrorIs(t, err, ErrInvalid)
	}
}

func TestMembershipPolicy(t *testing.T) {
	now := time.Now().UTC()
	token := "secret"
	host, err := RestorePlayer("host", "Alice", "", Host, HashCredential(token), now, "192.0.2.1")
	require.NoError(t, err)
	r, err := NewRoom("room", "ABC234", "PRIVATE", 3, host, now, "192.0.2.1")
	require.NoError(t, err)
	_, _, err = r.Join("h", "", "bad", "wrong", now, "198.51.100.99")
	require.ErrorIs(t, err, ErrUnauthorized)
	_, _, err = r.Join("g", "Bob", "guest", "", now, "198.51.100.99")
	require.NoError(t, err)
	require.Equal(t, Waiting, r.Status())
	_, _, err = r.Join("g2", "Mallory", "intruder", "", now, "198.51.100.99")
	require.ErrorIs(t, err, ErrFull)
	p, changed, err := r.Join("h", "", "host", token, now, "198.51.100.99")
	require.NoError(t, err)
	require.Equal(t, Host, p.Role())
	require.True(t, changed)
	require.Equal(t, Playing, r.Status())
	_, changed, err = r.Join("unused", "", "host", token, now, "198.51.100.99")
	require.NoError(t, err)
	require.False(t, changed)
	require.False(t, r.Leave("old"))
	require.True(t, r.Leave("guest"))
	require.False(t, r.Leave("guest"))
	require.Equal(t, Waiting, r.Status())
	_, _, err = r.Join("g", "Bob", "new", "", now, "198.51.100.99")
	require.NoError(t, err)
	require.Equal(t, Playing, r.Status())
}

func TestWaitingHostReconnect(t *testing.T) {
	now := time.Now()
	host, _ := RestorePlayer("host", "A", "", Host, HashCredential("token"), now, "192.0.2.1")
	r, _ := NewRoom("room", "ABC234", "PUBLIC", 1, host, now, "192.0.2.1")
	_, _, err := r.Join("h", "", "old", "token", now, "198.51.100.99")
	require.NoError(t, err)
	_, _, err = r.Join("h", "", "new", "token", now, "198.51.100.99")
	require.ErrorIs(t, err, ErrUnauthorized)
	require.True(t, r.Leave("old"))
	_, _, err = r.Join("h", "", "new", "token", now, "198.51.100.99")
	require.NoError(t, err)
	require.False(t, r.Leave("old"))
	_, ok := r.Member("new")
	require.True(t, ok)
}

func TestIPNormalizationAndRoomLimits(t *testing.T) {
	for input, want := range map[string]string{"192.0.2.1": "192.0.2.1", "::ffff:192.0.2.1": "192.0.2.1", "2001:0db8:0:0::1": "2001:db8::1"} {
		got, err := NormalizeIP(input)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
	for _, input := range []string{"", "not-an-ip", "192.0.2.1:1234", "192.0.2.1,192.0.2.2"} {
		_, err := NormalizeIP(input)
		require.ErrorIs(t, err, ErrInvalid)
	}
	require.ErrorIs(t, CheckRoomLimit(Waiting, 50), ErrWaitingRoomLimit)
	require.ErrorIs(t, CheckRoomLimit(Playing, 50), ErrActiveRoomLimit)
	require.NoError(t, CheckRoomLimit(Waiting, 49))
	require.NoError(t, CheckRoomLimit(Playing, 49))
	require.NoError(t, CheckRoomLimit(Playing, 0))
	require.NoError(t, CheckRoomLimit(Finished, 10))
}

func TestConnectedHostReconnect(t *testing.T) {
	now := time.Now()
	host, _ := RestorePlayer("host", "A", "", Host, HashCredential("token"), now, "192.0.2.1")
	r, _ := NewRoom("room", "ABC234", "PUBLIC", 1, host, now, "192.0.2.1")
	_, _, err := r.Join("h", "", "old-host", "token", now, "192.0.2.1")
	require.NoError(t, err)
	_, _, err = r.Join("g", "Guest", "guest", "", now, "192.0.2.2")
	require.NoError(t, err)
	require.True(t, r.Leave("old-host"))
	require.Equal(t, Waiting, r.Status())
	_, _, err = r.Join("h", "", "new-host", "wrong", now, "192.0.2.1")
	require.ErrorIs(t, err, ErrUnauthorized)
	_, _, err = r.Join("h", "", "new-host", "token", now, "192.0.2.1")
	require.NoError(t, err)
	require.Equal(t, Playing, r.Status())
	require.False(t, r.Leave("old-host"))
	require.Equal(t, Playing, r.Status())
}
