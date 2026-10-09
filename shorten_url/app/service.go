package app

import (
	"context"
	"errors"
	"time"

	"htqrcode/shorten_url/domain"

	"github.com/google/uuid"
)

var (
	ErrLiveURLLimit = errors.New("live short URL limit reached")
	ErrNotFound     = errors.New("short URL not found")
	ErrCodeConflict = errors.New("short code collision")
)

type Repository interface {
	Create(ctx context.Context, candidate domain.ShortURL, maxLive int) (domain.ShortURL, error)
	FindActiveByCode(ctx context.Context, code string, now time.Time) (domain.ShortURL, error)
	DeleteDeactivatedBefore(ctx context.Context, before time.Time) error
}

type Service struct {
	repo      Repository
	generator domain.CodeGenerator
	now       func() time.Time
}

func NewService(repo Repository, generator domain.CodeGenerator) *Service {
	if repo == nil || generator == nil {
		panic("short URL repository and generator are required")
	}
	return &Service{repo: repo, generator: generator, now: time.Now}
}

type CreateInput struct {
	UserID                  uuid.UUID
	Plan                    domain.Plan
	LongURL, IdempotencyKey string
}
type CreateOutput struct {
	URL      domain.ShortURL
	Existing bool
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	if input.UserID == uuid.Nil {
		input.Plan = domain.Free
	}
	limits, err := input.Plan.Limits()
	if err != nil {
		return CreateOutput{}, err
	}
	if input.UserID == uuid.Nil {
		limits.Lifetime = 7 * 24 * time.Hour
		limits.MaxLive = 0 // Anonymous creation has no per-user quota.
	}
	now := s.now().UTC()
	for attempt := 0; attempt < 10; attempt++ {
		code, err := s.generator.Generate()
		if err != nil {
			return CreateOutput{}, err
		}
		candidate, err := domain.NewShortURL(uuid.New(), input.UserID, input.LongURL, code, input.IdempotencyKey, now, limits.Lifetime)
		if err != nil {
			return CreateOutput{}, err
		}
		created, err := s.repo.Create(ctx, candidate, limits.MaxLive)
		if errors.Is(err, ErrCodeConflict) {
			continue
		}
		if err != nil {
			return CreateOutput{}, err
		}
		return CreateOutput{URL: created, Existing: created.ID() != candidate.ID()}, nil
	}
	return CreateOutput{}, errors.New("could not allocate a unique short code")
}

func (s *Service) Resolve(ctx context.Context, code string) (domain.ShortURL, error) {
	return s.repo.FindActiveByCode(ctx, code, s.now().UTC())
}

func (s *Service) Cleanup(ctx context.Context, now time.Time) error {
	return s.repo.DeleteDeactivatedBefore(ctx, now.UTC().Add(-365*24*time.Hour))
}
