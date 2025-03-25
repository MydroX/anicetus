package repository

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/pgxutil"
	"MydroX/anicetus/internal/iam/models"
	"MydroX/anicetus/pkg/logger"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	_, err := r.dbPool.Exec(ctx.StdContext(), r.queries.SaveSession(), session.UUID, session.UserId, session.RefreshToken, session.LastUsedAt, session.OS, session.Browser, session.BrowserVersion, session.IPv4Address, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)

		if pgErr.Code == pgerrcode.UniqueViolation {
			return errorsutil.New(errorsutil.ERROR_UNIQUE_VIOLATION, "session: unique violation", err)
		}

		r.logger.Zap.Sugar().Errorf("error creating user: %v", err)
		return errorsutil.New(errorsutil.ERROR_INTERNAL, "session: error during save", err)
	}

	return nil
}
