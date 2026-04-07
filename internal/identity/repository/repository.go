package repository

import (
	"context"
	"fmt"

	"MydroX/anicetus/internal/identity/models"
	"MydroX/anicetus/pkg/db/postgresql/pgx"
	"MydroX/anicetus/pkg/errs"
	"go.uber.org/zap"
)

var errorUserNotFound = "user not found"

type repository struct {
	logger  *zap.SugaredLogger
	dbPool  pgx.DBPool
	queries *IdentityQueries
}

// New is creating an interface for every method of the repository
func New(l *zap.SugaredLogger, dbPool pgx.DBPool) IdentityRepository {
	return &repository{
		logger:  l,
		dbPool:  dbPool,
		queries: &IdentityQueries{},
	}
}

func (r *repository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.dbPool.Exec(ctx, r.queries.CreateUser(), user.UUID, user.Username, user.Email, user.Password)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	var user models.User

	err := r.dbPool.QueryRow(ctx, r.queries.GetUserByUUID(), uuid).
		Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	return &user, nil
}

func (r *repository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	err := r.dbPool.QueryRow(ctx, r.queries.UpdateUser(), user.Username, user.Email, user.UUID).
		Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	return user, nil
}

func (r *repository) UpdatePassword(ctx context.Context, uuid, password string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.UpdatePassword(), password, uuid)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, errorUserNotFound, nil)
	}

	return nil
}

func (r *repository) UpdateEmail(ctx context.Context, uuid, email string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.UpdateEmail(), email, uuid)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return &errs.AppError{Code: errs.ErrorNotFound, Err: err, Message: errorUserNotFound}
	}

	return nil
}

func (r *repository) DeleteUser(ctx context.Context, uuid string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.DeleteUser(), uuid)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return &errs.AppError{Code: errs.ErrorNotFound, Err: err, Message: errorUserNotFound}
	}

	return nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	err := r.dbPool.QueryRow(ctx, r.queries.GetUserByEmail(), email).
		Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	return &user, nil
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User

	err := r.dbPool.QueryRow(ctx, r.queries.GetUserByUsername(), username).
		Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	return &user, nil
}

func (r *repository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetAllUsers())
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	var users []*models.User

	for rows.Next() {
		var user models.User

		err := rows.Scan(&user.UUID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, &errs.AppError{Code: errs.ErrorInternal, Err: fmt.Errorf("error scanning user: %v", err)}
		}

		users = append(users, &user)
	}

	return users, nil
}
