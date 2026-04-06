package repository

import (
	"context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/pgxutil"
	"MydroX/anicetus/internal/iam/models"
	"go.uber.org/zap"
)

func NewIAMStore(l *zap.SugaredLogger, dbPool pgxutil.DBPool) IamStore {
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
		return errorsutil.SQLErrorParser(err)
	}

	return nil
}
