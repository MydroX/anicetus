package repository

import (
	"MydroX/anicetus/internal/common/pgxutil"
	"go.uber.org/zap"
)

type repository struct {
	logger  *zap.SugaredLogger
	dbPool  pgxutil.DBPool
	queries *Queries
}
