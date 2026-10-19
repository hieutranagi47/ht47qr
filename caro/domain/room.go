package domain

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/netip"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrInvalid          = errors.New("invalid room input")
	ErrUnavailable      = errors.New("room unavailable")
	ErrFull             = errors.New("room full")
	ErrUnauthorized     = errors.New("membership not authorized")
	ErrWaitingRoomLimit = errors.New("waiting room limit reached")
	ErrActiveRoomLimit  = errors.New("active room limit reached")
)

// Alphabet excludes ambiguous letters and digits. Codes normalize to uppercase.
const (
	Alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	Lifetime = 24 * time.Hour
)

const (
	Host     = "PLAYER_1"
	Guest    = "PLAYER_2"
	Waiting  = "WAITING"
	Playing  = "PLAYING"
	Finished = "FINISHED"
)

type Player struct {
	id, name, socket, role, credentialHash, ip string
	joined                                     time.Time
}

func RestorePlayer(id, name, socket, role, credentialHash string, joined time.Time, ip string) (Player, error) {
	name = strings.TrimSpace(name)
	if ip != "" {
		var err error
		ip, err = NormalizeIP(ip)
		if err != nil {
			return Player{}, err
		}
	}
	if id == "" || !ValidName(name) || (role != Host && role != Guest) || (role == Host && len(credentialHash) != 64) || (role == Guest && credentialHash != "") || len(socket) > 100 || joined.IsZero() {
		return Player{}, ErrInvalid
	}
	return Player{id, name, socket, role, credentialHash, ip, joined.UTC()}, nil
}

func ValidName(name string) bool {
	return utf8.ValidString(name) && strings.TrimSpace(name) != "" && utf8.RuneCountInString(name) <= 50
}
func (p Player) ID() string             { return p.id }
func (p Player) Name() string           { return p.name }
func (p Player) Socket() string         { return p.socket }
func (p Player) Role() string           { return p.role }
func (p Player) CredentialHash() string { return p.credentialHash }
func (p Player) JoinedAt() time.Time    { return p.joined }

type Room struct {
	id, code, kind, status, ownerIP string
	maxGames                        int
	created, expires                time.Time
	players                         []Player
}

func NormalizeCode(code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 6 {
		return "", ErrInvalid
	}
	for _, c := range code {
		if !strings.ContainsRune(Alphabet, c) {
			return "", ErrInvalid
		}
	}
	return code, nil
}

func NewRoom(id, code, kind string, maxGames int, host Player, now time.Time, ownerIP string) (Room, error) {
	var err error
	ownerIP, err = NormalizeIP(ownerIP)
	if err != nil {
		return Room{}, err
	}
	if host.role != Host || host.socket != "" {
		return Room{}, ErrInvalid
	}
	return RestoreRoom(id, code, kind, Waiting, maxGames, now, now.Add(Lifetime), []Player{host}, ownerIP)
}

func RestoreRoom(id, code, kind, status string, maxGames int, created, expires time.Time, players []Player, ownerIP string) (Room, error) {
	code, err := NormalizeCode(code)
	if ownerIP != "" {
		normalized, ipErr := NormalizeIP(ownerIP)
		if ipErr != nil {
			return Room{}, ipErr
		}
		ownerIP = normalized
	}
	if err != nil || id == "" || (kind != "PUBLIC" && kind != "PRIVATE") || (maxGames != 1 && maxGames != 3 && maxGames != 5 && maxGames != 7) || (status != Waiting && status != Playing && status != Finished) || created.IsZero() || !expires.Equal(created.Add(Lifetime)) || len(players) < 1 || len(players) > 2 {
		return Room{}, ErrInvalid
	}
	roles := map[string]bool{}
	for _, p := range players {
		if roles[p.role] {
			return Room{}, ErrInvalid
		}
		roles[p.role] = true
	}
	if !roles[Host] {
		return Room{}, ErrInvalid
	}
	if status == Playing && (len(players) != 2 || players[0].socket == "" || players[1].socket == "") {
		return Room{}, ErrInvalid
	}
	return Room{id, code, kind, status, ownerIP, maxGames, created.UTC(), expires.UTC(), append([]Player(nil), players...)}, nil
}
func (r Room) ID() string                { return r.id }
func (r Room) Code() string              { return r.code }
func (r Room) Type() string              { return r.kind }
func (r Room) Status() string            { return r.status }
func (r Room) MaxGames() int             { return r.maxGames }
func (r Room) CreatedAt() time.Time      { return r.created }
func (r Room) ExpiresAt() time.Time      { return r.expires }
func (r Room) Players() []Player         { return append([]Player(nil), r.players...) }
func (r Room) Active(now time.Time) bool { return now.Before(r.expires) }
func (r Room) Joinable(now time.Time) bool {
	return r.Active(now) && r.status == Waiting && len(r.players) < 2
}

func (r Room) Member(socket string) (Player, bool) {
	for _, p := range r.players {
		if socket != "" && p.socket == socket {
			return p, true
		}
	}
	return Player{}, false
}

func HashCredential(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func NewCredential() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func RandomCode() (string, error) {
	b := make([]byte, 6)
	for i := range b {
		for {
			var v [1]byte
			if _, err := rand.Read(v[:]); err != nil {
				return "", err
			}
			if int(v[0]) < 256-256%len(Alphabet) {
				b[i] = Alphabet[int(v[0])%len(Alphabet)]
				break
			}
		}
	}
	return string(b), nil
}

// Join binds a reservation or adds a guest. A bound reservation cannot be stolen.
func (r *Room) Join(id, name, socket, token string, now time.Time, ip string) (Player, bool, error) {
	ip, err := NormalizeIP(ip)
	if err != nil {
		return Player{}, false, err
	}
	if !r.Active(now) || r.status == Finished {
		return Player{}, false, ErrUnavailable
	}
	if p, ok := r.Member(socket); ok {
		if p.ip != ip {
			return Player{}, false, ErrUnauthorized
		}
		return p, false, nil
	}
	if socket == "" || len(socket) > 100 {
		return Player{}, false, ErrInvalid
	}
	if token != "" {
		for i, p := range r.players {
			if p.role != Host {
				continue
			}
			hash := HashCredential(token)
			if subtle.ConstantTimeCompare([]byte(hash), []byte(p.credentialHash)) != 1 || p.socket != "" {
				return Player{}, false, ErrUnauthorized
			}
			r.players[i].socket = socket
			r.players[i].ip = ip
			r.startIfReady()
			return r.players[i], true, nil
		}
		return Player{}, false, ErrUnauthorized
	}
	if len(r.players) >= 2 || r.status != Waiting {
		return Player{}, false, ErrFull
	}
	p, err := RestorePlayer(id, name, socket, Guest, "", now, ip)
	if err != nil {
		return Player{}, false, err
	}
	r.players = append(r.players, p)
	r.startIfReady()
	return p, true, nil
}

func (r *Room) startIfReady() {
	if len(r.players) == 2 && r.players[0].socket != "" && r.players[1].socket != "" {
		r.status = Playing
	}
}

// A playing departure is terminal. Waiting hosts retain their private reservation.
func (r *Room) Leave(socket string) bool {
	for i, p := range r.players {
		if socket != "" && p.socket == socket {
			if r.status == Playing {
				r.status = Finished
			}
			if p.role == Host {
				r.players[i].socket = ""
			} else {
				r.players = append(r.players[:i], r.players[i+1:]...)
			}
			return true
		}
	}
	return false
}

// NormalizeIP merges IPv4-mapped IPv6 with IPv4 and strips connection-only zones.
func NormalizeIP(value string) (string, error) {
	ip, err := netip.ParseAddr(value)
	if err != nil {
		return "", ErrInvalid
	}
	return ip.Unmap().WithZone("").String(), nil
}
func (r Room) OwnerIP() string { return r.ownerIP }

// MaxRoomsPerIP caps the combined waiting and playing rooms for each IP.
const MaxRoomsPerIP = 50

// CheckRoomLimit checks other unexpired rooms owned or joined by the same IP.
func CheckRoomLimit(status string, count int64) error {
	if count < MaxRoomsPerIP {
		return nil
	}
	switch status {
	case Waiting:
		return ErrWaitingRoomLimit
	case Playing:
		return ErrActiveRoomLimit
	}
	return nil
}

func (p Player) IP() string { return p.ip }
func ValidateSettings(kind string, games int) error {
	if (kind != "PUBLIC" && kind != "PRIVATE") || (games != 1 && games != 3 && games != 5 && games != 7) {
		return ErrInvalid
	}
	return nil
}

// CreationSnapshot reproduces the original response without live socket bindings.
func (r Room) CreationSnapshot() (Room, error) {
	for _, p := range r.players {
		if p.role == Host {
			p.socket = ""
			p.ip = r.ownerIP
			return NewRoom(r.id, r.code, r.kind, r.maxGames, p, r.created, r.ownerIP)
		}
	}
	return Room{}, ErrInvalid
}

// RecoveryToken is stable for a private request key and a persisted random nonce.
// The database stores neither the key nor this token, only their hashes and nonce.
func RecoveryToken(key, nonce string) string {
	h := hmac.New(sha256.New, []byte(key))
	_, _ = h.Write([]byte("caro:host:v1:" + nonce))
	return hex.EncodeToString(h.Sum(nil))
}

// ParticipantIPs counts a room once for each creator or admitted player IP.
func (r Room) ParticipantIPs() []string {
	seen := map[string]bool{}
	ips := []string{}
	add := func(ip string) {
		if ip != "" && !seen[ip] {
			seen[ip] = true
			ips = append(ips, ip)
		}
	}
	add(r.ownerIP)
	for _, p := range r.players {
		add(p.ip)
	}
	return ips
}
