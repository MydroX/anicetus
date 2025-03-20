package repository

import (
	"MydroX/anicetus/internal/common/context"
	errorsutil "MydroX/anicetus/internal/common/errors"
	"MydroX/anicetus/internal/iam/models"
	"MydroX/anicetus/pkg/logger"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	logger *logger.Logger
	dbPool *pgxpool.Pool
}

func New(l *logger.Logger, dbPool *pgxpool.Pool) IamRepository {
	return &repository{
		logger: l,
		dbPool: dbPool,
	}
}

func (r *repository) SaveSession(ctx *context.AppContext, session *models.Session) *errorsutil.Err {
	query := `INSERT INTO sessions 
	(uuid, user_uuid, refresh_token, last_used_at, os, browser, browser_version, ipv4_addres, created_at, expires_at, ) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.dbPool.Exec(ctx.StdContext(), query, session.UUID, session.UserId, session.RefreshToken, session.LastUsedAt, session.OS, session.Browser, session.BrowserVersion, session.IPv4Address, session.CreatedAt, session.ExpiresAt)
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
