package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"MydroX/anicetus/pkg/db/postgresql/pgx"
	"MydroX/anicetus/pkg/errs"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func NewAudienceStore(l *zap.SugaredLogger, dbPool pgx.DBPool) AudienceStore {
	return &repository{
		logger:  l,
		dbPool:  dbPool,
		queries: &Queries{},
	}
}

func (r *repository) IsValidAudience(ctx context.Context, audience string) (bool, error) {
	var exists bool

	err := r.dbPool.QueryRow(ctx, r.queries.IsValidAudience(), audience).Scan(&exists)
	if err != nil {
		return false, errs.SQLErrorParser(err)
	}

	return exists, nil
}

func (r *repository) GetAllowedAudiences(ctx context.Context) ([]string, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetAllowedAudiences())
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}
	defer rows.Close()

	var audiences []string

	for rows.Next() {
		var audience string

		if err := rows.Scan(&audience); err != nil {
			return nil, errs.SQLErrorParser(err)
		}

		audiences = append(audiences, audience)
	}

	return audiences, nil
}

func (r *repository) GetUserAudiences(ctx context.Context, userUUID string) ([]string, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetUserAudiences(), userUUID)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}
	defer rows.Close()

	var audiences []string

	for rows.Next() {
		var audience string

		if err := rows.Scan(&audience); err != nil {
			return nil, errs.SQLErrorParser(err)
		}

		audiences = append(audiences, audience)
	}

	return audiences, nil
}

func (r *repository) RegisterAudience(ctx context.Context, audience string, metadata map[string]any) error {
	serviceName, ok := metadata["service_name"].(string)
	if !ok {
		serviceName = ""
	}

	description, ok := metadata["description"].(string)
	if !ok {
		description = ""
	}

	var permissionsJSON []byte

	if perms, ok := metadata["permissions"]; ok {
		var err error

		permissionsJSON, err = json.Marshal(perms)
		if err != nil {
			return fmt.Errorf("failed to marshal permissions: %w", err)
		}
	}

	uuid := uuid.Must(uuid.NewV7()).String()

	_, err := r.dbPool.Exec(ctx, r.queries.RegisterAudience(), uuid, audience, serviceName, description, permissionsJSON)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) RevokeAudience(ctx context.Context, audience string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.RevokeAudience(), audience)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "audience not found", nil)
	}

	return nil
}

func (r *repository) AssignAudienceToUser(ctx context.Context, userUUID, audience string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.AssignAudienceToUser(), userUUID, audience)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "audience not found or not active", nil)
	}

	return nil
}

func (r *repository) UnassignAudienceFromUser(ctx context.Context, userUUID, audience string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.UnassignAudienceFromUser(), userUUID, audience)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "audience assignment not found", nil)
	}

	return nil
}
