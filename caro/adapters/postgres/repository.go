package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"htqrcode/caro/adapters/postgres/dbmodels"
	"htqrcode/caro/app"
	"htqrcode/caro/domain"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		panic("Caro database required")
	}
	return &Repository{pool}
}

func (r *Repository) transaction(ctx context.Context, fn func(*dbmodels.Queries) error) error {
	var err error
	for range 8 {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = r.attempt(ctx, fn)
		var e *pgconn.PgError
		if !errors.As(err, &e) || (e.Code != "40001" && e.Code != "40P01") {
			return err
		}
	}
	return err
}

func (r *Repository) attempt(ctx context.Context, fn func(*dbmodels.Queries) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()
	if err := fn(dbmodels.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func stamp(t time.Time) pgtype.Timestamptz    { return pgtype.Timestamptz{Time: t.UTC(), Valid: true} }
func fromTime(t pgtype.Timestamptz) time.Time { return t.Time.UTC() }
func mapError(err error) error {
	var e *pgconn.PgError
	if errors.As(err, &e) && e.Code == "23505" {
		return app.ErrMembershipConflict
	}
	return err
}

func (r *Repository) Create(ctx context.Context, candidate app.Creation) (app.Creation, error) {
	var result app.Creation
	err := r.transaction(ctx, func(q *dbmodels.Queries) error {
		room := candidate.Room
		existing, err := findCreation(ctx, q, room.OwnerIP(), candidate.KeyHash, room.CreatedAt())
		if err == nil {
			if existing.Fingerprint != candidate.Fingerprint {
				return app.ErrIdempotencyConflict
			}
			result = existing
			return nil
		}
		if !errors.Is(err, app.ErrNotFound) {
			return err
		}
		count, err := q.InsertRoom(ctx, dbmodels.InsertRoomParams{ID: room.ID(), Code: room.Code(), RoomType: room.Type(), Status: room.Status(), MaxGames: int32(room.MaxGames()), CreatedAt: stamp(room.CreatedAt()), ExpiresAt: stamp(room.ExpiresAt()), OwnerIp: pgtype.Text{String: room.OwnerIP(), Valid: room.OwnerIP() != ""}, IdempotencyKeyHash: pgtype.Text{String: candidate.KeyHash, Valid: candidate.KeyHash != ""}, RequestFingerprint: pgtype.Text{String: candidate.Fingerprint, Valid: candidate.Fingerprint != ""}, RecoveryNonce: pgtype.Text{String: candidate.Nonce, Valid: candidate.Nonce != ""}})
		if err != nil {
			return err
		}
		if count == 0 {
			return app.ErrCodeConflict
		}
		if err := savePlayers(ctx, q, room); err != nil {
			return err
		}
		if err := checkParticipantLimits(ctx, q, room, room.CreatedAt()); err != nil {
			return err
		}
		row, err := q.RoomByCode(ctx, room.Code())
		if err != nil {
			return err
		}
		persisted, err := readRoom(ctx, q, row)
		if err != nil {
			return err
		}
		result = candidate
		result.Room = persisted
		return nil
	})
	return result, err
}

func findCreation(ctx context.Context, q *dbmodels.Queries, ip, keyHash string, now time.Time) (app.Creation, error) {
	row, err := q.RoomByCreationKey(ctx, dbmodels.RoomByCreationKeyParams{OwnerIp: pgtype.Text{String: ip, Valid: ip != ""}, IdempotencyKeyHash: pgtype.Text{String: keyHash, Valid: keyHash != ""}})
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Creation{}, app.ErrNotFound
	}
	if err != nil {
		return app.Creation{}, err
	}
	if !now.Before(fromTime(row.ExpiresAt)) {
		return app.Creation{}, app.ErrIdempotencyExpired
	}
	room, err := readRoom(ctx, q, row)
	if err != nil {
		return app.Creation{}, err
	}
	return app.Creation{Room: room, KeyHash: row.IdempotencyKeyHash.String, Fingerprint: row.RequestFingerprint.String, Nonce: row.RecoveryNonce.String}, nil
}

func (r *Repository) FindCreation(ctx context.Context, ip, keyHash string, now time.Time) (app.Creation, error) {
	var result app.Creation
	err := r.transaction(ctx, func(q *dbmodels.Queries) error {
		var err error
		result, err = findCreation(ctx, q, ip, keyHash, now)
		return err
	})
	return result, err
}

func savePlayers(ctx context.Context, q *dbmodels.Queries, r domain.Room) error {
	if err := q.DeletePlayers(ctx, r.ID()); err != nil {
		return err
	}
	for _, p := range r.Players() {
		err := q.InsertPlayer(ctx, dbmodels.InsertPlayerParams{ID: p.ID(), RoomID: r.ID(), DisplayName: p.Name(), SocketID: pgtype.Text{String: p.Socket(), Valid: p.Socket() != ""}, PlayerRole: p.Role(), CredentialHash: p.CredentialHash(), JoinedAt: stamp(p.JoinedAt()), ParticipantIp: pgtype.Text{String: p.IP(), Valid: p.IP() != ""}})
		if err != nil {
			return mapError(err)
		}
	}
	return nil
}

func readRoom(ctx context.Context, q *dbmodels.Queries, row dbmodels.CaroRoom) (domain.Room, error) {
	rows, err := q.Players(ctx, row.ID)
	if err != nil {
		return domain.Room{}, err
	}
	players := make([]domain.Player, 0, len(rows))
	for _, p := range rows {
		v, err := domain.RestorePlayer(p.ID, p.DisplayName, p.SocketID.String, p.PlayerRole, p.CredentialHash, fromTime(p.JoinedAt), p.ParticipantIp.String)
		if err != nil {
			return domain.Room{}, err
		}
		players = append(players, v)
	}
	return domain.RestoreRoom(row.ID, row.Code, row.RoomType, row.Status, int(row.MaxGames), fromTime(row.CreatedAt), fromTime(row.ExpiresAt), players, row.OwnerIp.String)
}

func (r *Repository) FindActive(ctx context.Context, code string, now time.Time) (domain.Room, error) {
	var result domain.Room
	err := r.transaction(ctx, func(q *dbmodels.Queries) error {
		row, err := q.RoomByCode(ctx, code)
		if errors.Is(err, pgx.ErrNoRows) {
			return app.ErrNotFound
		}
		if err != nil {
			return err
		}
		if !now.Before(fromTime(row.ExpiresAt)) {
			return app.ErrNotFound
		}
		result, err = readRoom(ctx, q, row)
		return err
	})
	return result, err
}

func (r *Repository) ListWaiting(ctx context.Context, now time.Time) ([]domain.Room, error) {
	result := []domain.Room{}
	err := r.transaction(ctx, func(q *dbmodels.Queries) error {
		rows, err := q.PublicWaiting(ctx, stamp(now))
		if err != nil {
			return err
		}
		result = []domain.Room{}
		for _, row := range rows {
			room, err := readRoom(ctx, q, row)
			if err != nil {
				return err
			}
			result = append(result, room)
		}
		return nil
	})
	return result, err
}

func (r *Repository) Update(ctx context.Context, code string, now time.Time, fn func(*domain.Room) error) (domain.Room, error) {
	var result domain.Room
	err := r.transaction(ctx, func(q *dbmodels.Queries) error {
		row, err := q.LockRoom(ctx, code)
		if errors.Is(err, pgx.ErrNoRows) {
			return app.ErrNotFound
		}
		if err != nil {
			return err
		}
		if !now.Before(fromTime(row.ExpiresAt)) {
			return domain.ErrUnavailable
		}
		result, err = readRoom(ctx, q, row)
		if err != nil {
			return err
		}
		beforeIPs := result.ParticipantIPs()
		if err := fn(&result); err != nil {
			return err
		}
		if err := q.SaveStatus(ctx, dbmodels.SaveStatusParams{ID: result.ID(), Status: result.Status()}); err != nil {
			return err
		}
		if err := savePlayers(ctx, q, result); err != nil {
			return err
		}
		if row.Status != result.Status() || addedParticipantIP(beforeIPs, result.ParticipantIPs()) {
			return checkParticipantLimits(ctx, q, result, now)
		}
		return nil
	})
	return result, err
}

func (r *Repository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	var count int64
	err := r.transaction(ctx, func(q *dbmodels.Queries) error {
		var err error
		count, err = q.DeleteExpired(ctx, stamp(now))
		return err
	})
	return count, err
}

func (r *Repository) ResetConnections(ctx context.Context) error {
	return r.transaction(ctx, func(q *dbmodels.Queries) error {
		if err := q.ResetRoomStatus(ctx); err != nil {
			return err
		}
		if err := q.ResetGuests(ctx); err != nil {
			return err
		}
		return q.ResetSockets(ctx)
	})
}

var _ app.Repository = (*Repository)(nil)

// Predicate checks and membership writes commit together under the SQLite writer
// lock or PostgreSQL SERIALIZABLE isolation, including joins to different rooms.
func checkParticipantLimits(ctx context.Context, q *dbmodels.Queries, room domain.Room, now time.Time) error {
	if room.Status() != domain.Waiting && room.Status() != domain.Playing {
		return nil
	}
	for _, ip := range room.ParticipantIPs() {
		count, err := q.CountParticipantRooms(ctx, dbmodels.CountParticipantRoomsParams{Ip: pgtype.Text{String: ip, Valid: ip != ""}, Cutoff: stamp(now), ExcludeID: room.ID()})
		if err != nil {
			return err
		}
		if err := domain.CheckRoomLimit(room.Status(), count); err != nil {
			return err
		}
	}
	return nil
}

func addedParticipantIP(before, after []string) bool {
	seen := map[string]bool{}
	for _, ip := range before {
		seen[ip] = true
	}
	for _, ip := range after {
		if !seen[ip] {
			return true
		}
	}
	return false
}
