//nolint:all
package repository

import (
	"MydroX/anicetus/internal/common/pgxutil"

	"go.uber.org/zap"
)

func NewAudienceStore(l *zap.SugaredLogger, dbPool pgxutil.DBPool) AudienceStore {
	return &repository{
		logger:  l,
		dbPool:  dbPool,
		queries: &Queries{},
	}
}
func (r *repository) IsValidAudience(audience string) (bool, error) {
	panic("not implemented") // TODO: Implement
}

func (r *repository) GetAllowedAudiences() ([]string, error) {
	panic("not implemented") // TODO: Implement
}

func (r *repository) RegisterAudience(audience string, metadata map[string]any) error {
	panic("not implemented") // TODO: Implement
}

func (r *repository) RevokeAudience(audience string) error {
	panic("not implemented") // TODO: Implement
}
