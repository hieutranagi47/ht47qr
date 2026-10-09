package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"htqrcode/shorten_url/adapters/postgres/dbmodels"
	"htqrcode/shorten_url/app"
)

func (r *Repository) GetMetadata(ctx context.Context, id uuid.UUID) (app.Metadata, error) {
	value, err := dbmodels.New(r.pool).GetMetadata(ctx, id.String())
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Metadata{}, app.ErrMetadataUnavailable
	}
	if err != nil {
		return app.Metadata{}, err
	}
	return app.Metadata{Title: value.Title, Description: value.Description, ImageURL: value.ImageUrl, FetchedAt: time.Unix(0, value.FetchedAt.Int64).UTC()}, nil
}

func (r *Repository) ClaimMetadata(ctx context.Context, now time.Time) (app.MetadataJob, error) {
	q := dbmodels.New(r.pool)
	if err := q.ExpireMetadataAttempts(ctx, now.UnixNano()); err != nil {
		return app.MetadataJob{}, err
	}
	value, err := q.ClaimMetadata(ctx, dbmodels.ClaimMetadataParams{Now: now.UnixNano(), LeaseUntil: now.Add(30 * time.Second).UnixNano()})
	if errors.Is(err, pgx.ErrNoRows) {
		return app.MetadataJob{}, app.ErrMetadataUnavailable
	}
	if err != nil {
		return app.MetadataJob{}, err
	}
	id, err := uuid.Parse(value.LinkID)
	if err != nil {
		return app.MetadataJob{}, err
	}
	target, err := q.GetMetadataJobURL(ctx, value.LinkID)
	return app.MetadataJob{LinkID: id, LongURL: target, Attempt: value.Attempts}, err
}

func (r *Repository) CompleteMetadata(ctx context.Context, job app.MetadataJob, m app.Metadata) error {
	return dbmodels.New(r.pool).CompleteMetadata(ctx, dbmodels.CompleteMetadataParams{
		LinkID: job.LinkID.String(), Attempts: job.Attempt, Title: m.Title, Description: m.Description, ImageUrl: m.ImageURL,
		FetchedAt: sql.NullInt64{Int64: m.FetchedAt.UnixNano(), Valid: true},
	})
}

func (r *Repository) FailMetadata(ctx context.Context, job app.MetadataJob, retryAt time.Time) error {
	return dbmodels.New(r.pool).FailMetadata(ctx, dbmodels.FailMetadataParams{LinkID: job.LinkID.String(), Attempts: job.Attempt, NextAttemptAt: retryAt.UnixNano()})
}
