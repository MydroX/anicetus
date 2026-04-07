package repository

import (
	"context"

	"MydroX/anicetus/internal/authentication/models"
	"MydroX/anicetus/pkg/db/postgresql/pgx"
	"MydroX/anicetus/pkg/errs"
	"go.uber.org/zap"
)

type repository struct {
	logger  *zap.SugaredLogger
	dbPool  pgx.DBPool
	queries *SessionQueries
}

func New(l *zap.SugaredLogger, dbPool pgx.DBPool) SessionStore {
	return &repository{
		logger:  l,
		dbPool:  dbPool,
		queries: &SessionQueries{},
	}
}

func (r *repository) SaveSession(ctx context.Context, session *models.Session) error {
	_, err := r.dbPool.Exec(
		ctx,
		r.queries.SaveSession(),
		session.UUID,
		session.UserID,
		session.RefreshToken,
		session.LastUsedAt,
		session.OS,
		session.OSVersion,
		session.Browser,
		session.BrowserVersion,
		session.IPv4Address,
		session.CreatedAt,
		session.ExpiresAt,
	)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) GetSessionByUUID(ctx context.Context, uuid string) (*models.Session, error) {
	var session models.Session

	err := r.dbPool.QueryRow(ctx, r.queries.GetSessionByUUID(), uuid).
		Scan(
			&session.UUID,
			&session.UserID,
			&session.RefreshToken,
			&session.LastUsedAt,
			&session.OS,
			&session.OSVersion,
			&session.Browser,
			&session.BrowserVersion,
			&session.IPv4Address,
			&session.CreatedAt,
			&session.ExpiresAt,
		)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	return &session, nil
}

func (r *repository) DeleteSession(ctx context.Context, uuid string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.DeleteSession(), uuid)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "session not found", nil)
	}

	return nil
}

func (r *repository) DeleteAllUserSessions(ctx context.Context, userUUID string) error {
	_, err := r.dbPool.Exec(ctx, r.queries.DeleteAllUserSessions(), userUUID)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) UpdateSessionLastUsed(ctx context.Context, uuid string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.UpdateSessionLastUsed(), uuid)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "session not found", nil)
	}

	return nil
}
