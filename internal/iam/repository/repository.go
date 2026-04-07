package repository

import (
	"MydroX/anicetus/pkg/db/postgresql/pgx"
	"go.uber.org/zap"
)

type repository struct {
	logger  *zap.SugaredLogger
	dbPool  pgx.DBPool
	queries *Queries
}
