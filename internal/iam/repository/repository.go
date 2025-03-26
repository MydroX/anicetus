package repository

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/pgxutil"
	"MydroX/anicetus/internal/iam/models"
	"MydroX/anicetus/pkg/logger"
)

type repository struct {
	logger  *logger.Logger
	dbPool  pgxutil.DBPool
	queries *Queries
}

func New(l *logger.Logger, dbPool pgxutil.DBPool) IamRepository {
	return &repository{
		logger:  l,
		dbPool:  dbPool,
		queries: &Queries{},
	}
}

func (r *repository) SaveSession(ctx *context.AppContext, session *models.Session) error {
	_, err := r.dbPool.Exec(ctx.StdContext(), r.queries.SaveSession(), session.UUID, session.UserId, session.RefreshToken, session.LastUsedAt, session.OS, session.OSVersion, session.Browser, session.BrowserVersion, session.IPv4Address, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		return errorsutil.SQLErrorParser(err)
	}

	return nil
}
