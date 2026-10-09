package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"htqrcode/common"
	"htqrcode/shorten_url/adapters/db/dbmodels"
	"htqrcode/shorten_url/app"
	"htqrcode/shorten_url/domain"

	"github.com/google/uuid"
)

type Repository struct{ pool *sql.DB }

func NewRepository(pool *sql.DB) *Repository {
	if pool == nil {
		panic("short URL database pool is required")
	}
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, candidate domain.ShortURL, maxLive int) (domain.ShortURL, error) {
	var result domain.ShortURL
	err := common.UpdateInSQLiteTx(ctx, r.pool, func(ctx context.Context, tx *sql.Conn) error {
		queries := dbmodels.New(tx)
		userID := toDBUUID(candidate.UserID())

		existing, err := queries.GetLinkByUserAndIdempotency(ctx, dbmodels.GetLinkByUserAndIdempotencyParams{
			UserID: userID, IdempotencyKey: candidate.IdempotencyKey(),
		})
		if err == nil {
			result, err = toDomain(existing)
			return err
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if maxLive > 0 {
			count, err := queries.CountLiveLinksByUser(ctx, dbmodels.CountLiveLinksByUserParams{UserID: userID, ExpiresAt: candidate.CreatedAt().UnixNano()})
			if err != nil {
				return err
			}
			if count >= int64(maxLive) {
				return app.ErrLiveURLLimit
			}
		}

		inserted, err := queries.InsertLink(ctx, dbmodels.InsertLinkParams{
			ID:             toDBUUID(candidate.ID()),
			UserID:         userID,
			LongUrl:        candidate.LongURL(),
			ShortCode:      candidate.Code(),
			IdempotencyKey: candidate.IdempotencyKey(),
			CreatedAt:      candidate.CreatedAt().UnixNano(),
			ExpiresAt:      candidate.ExpiresAt().UnixNano(),
		})
		if errors.Is(err, sql.ErrNoRows) {
			existing, lookupErr := queries.GetLinkByUserAndIdempotency(ctx, dbmodels.GetLinkByUserAndIdempotencyParams{UserID: userID, IdempotencyKey: candidate.IdempotencyKey()})
			if lookupErr == nil {
				result, lookupErr = toDomain(existing)
				return lookupErr
			}
			if errors.Is(lookupErr, sql.ErrNoRows) {
				return app.ErrCodeConflict
			}
			return lookupErr
		}
		if err != nil {
			return err
		}
		if err := queries.EnqueueMetadata(ctx, inserted.ID); err != nil {
			return err
		}
		result, err = toDomain(inserted)
		return err
	})
	return result, err
}

func (r *Repository) FindActiveByCode(ctx context.Context, code string, now time.Time) (domain.ShortURL, error) {
	value, err := dbmodels.New(r.pool).GetActiveLinkByCode(ctx, dbmodels.GetActiveLinkByCodeParams{ShortCode: code, ExpiresAt: now.UnixNano()})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ShortURL{}, app.ErrNotFound
	}
	if err != nil {
		return domain.ShortURL{}, err
	}
	return toDomain(value)
}

func (r *Repository) DeleteDeactivatedBefore(ctx context.Context, before time.Time) error {
	return dbmodels.New(r.pool).DeleteOldLinks(ctx, sql.NullInt64{Int64: before.UnixNano(), Valid: true})
}

func toDBUUID(value uuid.UUID) string { return value.String() }

func toDomain(value dbmodels.ShortenUrlLink) (domain.ShortURL, error) {
	id, err := uuid.Parse(value.ID)
	if err != nil {
		return domain.ShortURL{}, err
	}
	userID, err := uuid.Parse(value.UserID)
	if err != nil {
		return domain.ShortURL{}, err
	}
	var deactivatedAt *time.Time
	if value.DeactivatedAt.Valid {
		t := time.Unix(0, value.DeactivatedAt.Int64).UTC()
		deactivatedAt = &t
	}
	return domain.RestoreShortURL(id, userID, value.LongUrl, value.ShortCode,
		value.IdempotencyKey, time.Unix(0, value.CreatedAt), time.Unix(0, value.ExpiresAt), deactivatedAt)
}
