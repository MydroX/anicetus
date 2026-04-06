package repository

import (
	"context"
	"fmt"

	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/pgxutil"
	"MydroX/anicetus/internal/users/models"
	"go.uber.org/zap"
)

var errorUserNotFound = "user not found"

type repository struct {
	logger  *zap.SugaredLogger
	dbPool  pgxutil.DBPool
	queries *UsersQueries
}

// New is creating an interface for every method of the repository
func New(l *zap.SugaredLogger, dbPool pgxutil.DBPool) UsersRepository {
	return &repository{
		logger:  l,
		dbPool:  dbPool,
		queries: &UsersQueries{},
	}
}

func (r *repository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.dbPool.Exec(ctx, r.queries.CreateUser(), user.UUID, user.Username, user.Email, user.Password, user.Role)
	if err != nil {
		return errorsutil.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	var user models.User

	err := r.dbPool.QueryRow(ctx, r.queries.GetUserByUUID(), uuid).
		Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errorsutil.SQLErrorParser(err)
	}

	return &user, nil
}

func (r *repository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	err := r.dbPool.QueryRow(ctx, r.queries.UpdateUser(), user.Username, user.Email, user.Role, user.UUID).
		Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errorsutil.SQLErrorParser(err)
	}

	return user, nil
}

func (r *repository) UpdatePassword(ctx context.Context, uuid, password string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.UpdatePassword(), password, uuid)
	if err != nil {
		return errorsutil.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errorsutil.New(errorsutil.ErrorNotFound, errorUserNotFound, nil)
	}

	return nil
}

func (r *repository) UpdateEmail(ctx context.Context, uuid, email string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.UpdateEmail(), email, uuid)
	if err != nil {
		return errorsutil.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return &errorsutil.AppError{Code: errorsutil.ErrorNotFound, Err: err, Message: errorUserNotFound}
	}

	return nil
}

func (r *repository) DeleteUser(ctx context.Context, uuid string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.DeleteUser(), uuid)
	if err != nil {
		return errorsutil.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return &errorsutil.AppError{Code: errorsutil.ErrorNotFound, Err: err, Message: errorUserNotFound}
	}

	return nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	err := r.dbPool.QueryRow(ctx, r.queries.GetUserByEmail(), email).
		Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errorsutil.SQLErrorParser(err)
	}

	return &user, nil
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User

	err := r.dbPool.QueryRow(ctx, r.queries.GetUserByUsername(), username).
		Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errorsutil.SQLErrorParser(err)
	}

	return &user, nil
}

func (r *repository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetAllUsers())
	if err != nil {
		return nil, errorsutil.SQLErrorParser(err)
	}

	var users []*models.User

	for rows.Next() {
		var user models.User

		err := rows.Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, &errorsutil.AppError{Code: errorsutil.ErrorInternal, Err: fmt.Errorf("error scanning user: %v", err)}
		}

		users = append(users, &user)
	}

	return users, nil
}
