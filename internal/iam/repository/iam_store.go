package repository

import (
	"context"

	"MydroX/anicetus/internal/iam/models"
	"MydroX/anicetus/pkg/db/postgresql/pgx"
	"MydroX/anicetus/pkg/errs"
	"go.uber.org/zap"
)

func NewIAMStore(l *zap.SugaredLogger, dbPool pgx.DBPool) IamStore {
	return &repository{
		logger:  l,
		dbPool:  dbPool,
		queries: &Queries{},
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
