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

type repository struct {
	logger  *zap.SugaredLogger
	dbPool  pgx.DBPool
	queries *ServiceQueries
}

func New(l *zap.SugaredLogger, dbPool pgx.DBPool) ServiceStore {
	return &repository{
		logger:  l,
		dbPool:  dbPool,
		queries: &ServiceQueries{},
	}
}

func (r *repository) IsValidService(ctx context.Context, audience string) (bool, error) {
	var exists bool

	err := r.dbPool.QueryRow(ctx, r.queries.IsValidService(), audience).Scan(&exists)
	if err != nil {
		return false, errs.SQLErrorParser(err)
	}

	return exists, nil
}

func (r *repository) GetAllowedServices(ctx context.Context) ([]string, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetAllowedServices())
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}
	defer rows.Close()

	var services []string

	for rows.Next() {
		var service string

		if err := rows.Scan(&service); err != nil {
			return nil, errs.SQLErrorParser(err)
		}

		services = append(services, service)
	}

	return services, nil
}

func (r *repository) GetUserServices(ctx context.Context, userUUID string) ([]string, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetUserServices(), userUUID)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}
	defer rows.Close()

	var services []string

	for rows.Next() {
		var service string

		if err := rows.Scan(&service); err != nil {
			return nil, errs.SQLErrorParser(err)
		}

		services = append(services, service)
	}

	return services, nil
}

func (r *repository) RegisterService(ctx context.Context, audience string, metadata map[string]any) error {
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

	serviceUUID := uuid.Must(uuid.NewV7()).String()

	_, err := r.dbPool.Exec(ctx, r.queries.RegisterService(), serviceUUID, audience, serviceName, description, permissionsJSON)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) RevokeService(ctx context.Context, audience string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.RevokeService(), audience)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "service not found", nil)
	}

	return nil
}

func (r *repository) AssignServiceToUser(ctx context.Context, userUUID, audience string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.AssignServiceToUser(), userUUID, audience)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "service not found or not active", nil)
	}

	return nil
}

func (r *repository) UnassignServiceFromUser(ctx context.Context, userUUID, audience string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.UnassignServiceFromUser(), userUUID, audience)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "service assignment not found", nil)
	}

	return nil
}
