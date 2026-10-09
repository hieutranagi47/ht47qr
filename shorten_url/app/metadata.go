package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"htqrcode/common/log"
)

var ErrMetadataUnavailable = errors.New("metadata unavailable")

// Metadata is a best-effort snapshot. ImageURL references the source image;
// no image bytes are fetched or stored.
type Metadata struct {
	Title, Description, ImageURL string
	FetchedAt                    time.Time
}

func (m Metadata) Available() bool { return m.Title != "" || m.Description != "" || m.ImageURL != "" }

type MetadataJob struct {
	LinkID  uuid.UUID
	LongURL string
	Attempt int64
}

type MetadataRepository interface {
	GetMetadata(context.Context, uuid.UUID) (Metadata, error)
	ClaimMetadata(context.Context, time.Time) (MetadataJob, error)
	CompleteMetadata(context.Context, MetadataJob, Metadata) error
	FailMetadata(context.Context, MetadataJob, time.Time) error
}
type MetadataFetcher interface {
	Fetch(context.Context, string) (Metadata, error)
}

type MetadataService struct {
	repo    MetadataRepository
	fetcher MetadataFetcher
}

func NewMetadataService(repo MetadataRepository, fetcher MetadataFetcher) *MetadataService {
	if repo == nil || fetcher == nil {
		panic("metadata repository and fetcher are required")
	}
	return &MetadataService{repo: repo, fetcher: fetcher}
}

func (s *MetadataService) Get(ctx context.Context, id uuid.UUID) (Metadata, error) {
	return s.repo.GetMetadata(ctx, id)
}

// ProcessNext handles one durable job. Claims expire after 30 seconds; each
// fetch has a shorter deadline, and failed jobs get at most three attempts.
func (s *MetadataService) ProcessNext(ctx context.Context) (bool, error) {
	job, err := s.repo.ClaimMetadata(ctx, time.Now().UTC())
	if errors.Is(err, ErrMetadataUnavailable) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	snapshot, fetchErr := s.fetcher.Fetch(fetchCtx, job.LongURL)
	cancel()
	if ctx.Err() != nil {
		return true, ctx.Err()
	}
	if fetchErr != nil {
		// Avoid logging target URLs, which may include credentials or private tokens.
		log.FromContext(ctx).Debug("metadata fetch failed", "link_id", job.LinkID, "attempt", job.Attempt)
		return true, s.repo.FailMetadata(ctx, job, time.Now().UTC().Add(time.Duration(job.Attempt)*time.Minute))
	}
	snapshot.FetchedAt = time.Now().UTC()
	return true, s.repo.CompleteMetadata(ctx, job, snapshot)
}

func (s *MetadataService) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		processed, err := s.ProcessNext(ctx)
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			log.FromContext(ctx).Warn("metadata job failed", "error", err)
		}
		if processed && err == nil {
			continue
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
		}
	}
	return nil
}
