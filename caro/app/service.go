package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"htqrcode/caro/domain"
)

var (
	ErrNotFound            = errors.New("room not found")
	ErrCodeConflict        = errors.New("room code conflict")
	ErrCollisionExhausted  = errors.New("room code attempts exhausted")
	ErrMembershipConflict  = errors.New("connection already belongs to a room")
	ErrIdempotencyConflict = errors.New("idempotency key used with different inputs")
	ErrIdempotencyExpired  = errors.New("idempotent room expired")
)

type Repository interface {
	Create(context.Context, Creation) (Creation, error)
	FindCreation(context.Context, string, string, time.Time) (Creation, error)
	FindActive(context.Context, string, time.Time) (domain.Room, error)
	ListWaiting(context.Context, time.Time) ([]domain.Room, error)
	Update(context.Context, string, time.Time, func(*domain.Room) error) (domain.Room, error)
	DeleteExpired(context.Context, time.Time) (int64, error)
	ResetConnections(context.Context) error
}

type Service struct {
	repo Repository
	now  func() time.Time
	code func() (string, error)
}

func NewService(repo Repository, now func() time.Time, code func() (string, error)) *Service {
	return &Service{repo, now, code}
}

type (
	CreateInput struct {
		DisplayName, RoomType, OwnerIP, IdempotencyKey string
		MaxGames                                       int
	}
	Created struct {
		Room              domain.Room
		HostID, HostToken string
	}
)

// Creation contains immutable retry metadata, separate from the public response.
type Creation struct {
	Room                        domain.Room
	KeyHash, Fingerprint, Nonce string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (Created, error) {
	ownerIP, err := domain.NormalizeIP(in.OwnerIP)
	if err != nil {
		return Created{}, err
	}
	key, err := uuid.Parse(in.IdempotencyKey)
	if err != nil || key.Version() != 4 || key.Variant() != uuid.RFC4122 {
		return Created{}, domain.ErrInvalid
	}
	in.IdempotencyKey = key.String()
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	if !domain.ValidName(in.DisplayName) {
		return Created{}, domain.ErrInvalid
	}
	if err := domain.ValidateSettings(in.RoomType, in.MaxGames); err != nil {
		return Created{}, err
	}
	body, _ := json.Marshal(struct {
		Name, Type string
		Games      int
	}{in.DisplayName, in.RoomType, in.MaxGames})
	sum := sha256.Sum256(body)
	fingerprint := hex.EncodeToString(sum[:])
	keyHash := domain.HashCredential(in.IdempotencyKey)
	now := s.now().UTC()
	existing, err := s.repo.FindCreation(ctx, ownerIP, keyHash, now)
	if err == nil {
		return creationResult(existing, in.IdempotencyKey, fingerprint)
	}
	if !errors.Is(err, ErrNotFound) {
		return Created{}, err
	}
	nonce, err := domain.NewCredential()
	if err != nil {
		return Created{}, err
	}
	token := domain.RecoveryToken(in.IdempotencyKey, nonce)
	host, err := domain.RestorePlayer(uuid.NewString(), in.DisplayName, "", domain.Host, domain.HashCredential(token), now, ownerIP)
	if err != nil {
		return Created{}, err
	}
	for range 8 {
		code, err := s.code()
		if err != nil {
			return Created{}, err
		}
		room, err := domain.NewRoom(uuid.NewString(), code, in.RoomType, in.MaxGames, host, now, ownerIP)
		if err != nil {
			return Created{}, err
		}
		creation, err := s.repo.Create(ctx, Creation{Room: room, KeyHash: keyHash, Fingerprint: fingerprint, Nonce: nonce})
		if errors.Is(err, ErrCodeConflict) {
			continue
		}
		if err != nil {
			return Created{}, err
		}
		return creationResult(creation, in.IdempotencyKey, fingerprint)
	}
	return Created{}, ErrCollisionExhausted
}

func creationResult(c Creation, key, fingerprint string) (Created, error) {
	if c.Fingerprint != fingerprint {
		return Created{}, ErrIdempotencyConflict
	}
	snapshot, err := c.Room.CreationSnapshot()
	if err != nil {
		return Created{}, err
	}
	return Created{Room: snapshot, HostID: snapshot.Players()[0].ID(), HostToken: domain.RecoveryToken(key, c.Nonce)}, nil
}

func (s *Service) Get(ctx context.Context, code string) (domain.Room, error) {
	code, err := domain.NormalizeCode(code)
	if err != nil {
		return domain.Room{}, err
	}
	room, err := s.repo.FindActive(ctx, code, s.now().UTC())
	if err == nil && !room.Active(s.now().UTC()) {
		return domain.Room{}, ErrNotFound
	}
	return room, err
}

func (s *Service) List(ctx context.Context) ([]domain.Room, error) {
	return s.repo.ListWaiting(ctx, s.now().UTC())
}

type Joined struct {
	Room    domain.Room
	Player  domain.Player
	Changed bool
}

func (s *Service) Join(ctx context.Context, code, name, socket, token, ip string) (Joined, error) {
	code, err := domain.NormalizeCode(code)
	if err != nil {
		return Joined{}, err
	}
	ip, err = domain.NormalizeIP(ip)
	if err != nil {
		return Joined{}, err
	}
	now := s.now().UTC()
	var p domain.Player
	var changed bool
	room, err := s.repo.Update(ctx, code, now, func(r *domain.Room) error {
		var err error
		p, changed, err = r.Join(uuid.NewString(), name, socket, token, s.now().UTC(), ip)
		return err
	})
	return Joined{room, p, changed}, err
}

func (s *Service) Leave(ctx context.Context, code, socket string) (Joined, error) {
	code, err := domain.NormalizeCode(code)
	if err != nil {
		return Joined{}, err
	}
	var changed bool
	room, err := s.repo.Update(ctx, code, s.now().UTC(), func(r *domain.Room) error { changed = r.Leave(socket); return nil })
	return Joined{Room: room, Changed: changed}, err
}

// Peer verifies persisted membership and expiry on every relay.
func (s *Service) Peer(ctx context.Context, code, socket string) (string, error) {
	room, err := s.Get(ctx, code)
	if err != nil {
		return "", err
	}
	if _, ok := room.Member(socket); !ok || room.Status() != domain.Playing {
		return "", domain.ErrUnauthorized
	}
	for _, p := range room.Players() {
		if p.Socket() != "" && p.Socket() != socket {
			return p.Socket(), nil
		}
	}
	return "", domain.ErrUnavailable
}

func (s *Service) ResetConnections(ctx context.Context) error { return s.repo.ResetConnections(ctx) }
