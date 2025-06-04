package repository

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/pgxutil"
	"MydroX/anicetus/internal/users/models"
	"fmt"

	"go.uber.org/zap"
)

var (
	errorUserNotFound = "user not found"
)

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

func (r *repository) CreateUser(ctx *context.AppContext, user *models.User) error {
	_, err := r.dbPool.Exec(ctx.StdContext(), r.queries.CreateUser(), user.UUID, user.Username, user.Email, user.Password, user.Role)
	if err != nil {
		return errorsutil.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) GetUserByUUID(ctx *context.AppContext, uuid string) (*models.User, error) {
	var user models.User
	err := r.dbPool.QueryRow(ctx.StdContext(), r.queries.GetUserByUUID(), uuid).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errorsutil.SQLErrorParser(err)
	}

	return &user, nil
}

func (r *repository) UpdateUser(ctx *context.AppContext, user *models.User) (*models.User, error) {
	err := r.dbPool.QueryRow(ctx.StdContext(), r.queries.UpdateUser(), user.Username, user.Email, user.Role, user.UUID).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errorsutil.SQLErrorParser(err)
	}

	return user, nil
}

func (r *repository) UpdatePassword(ctx *context.AppContext, uuid, password string) error {
	res, err := r.dbPool.Exec(ctx.StdContext(), r.queries.UpdatePassword(), password, uuid)
	if err != nil {
		return errorsutil.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errorsutil.New(errorsutil.ErrorNotFound, errorUserNotFound, nil)
	}

	return nil
}

func (r *repository) UpdateEmail(ctx *context.AppContext, uuid, email string) error {
	res, err := r.dbPool.Exec(ctx.StdContext(), r.queries.UpdateEmail(), email, uuid)
	if err != nil {
		return errorsutil.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return &errorsutil.AppError{Code: errorsutil.ErrorNotFound, Err: err, Message: errorUserNotFound}
	}

	return nil
}

func (r *repository) DeleteUser(ctx *context.AppContext, uuid string) error {
	res, err := r.dbPool.Exec(ctx.StdContext(), r.queries.DeleteUser(), uuid)
	if err != nil {
		return errorsutil.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return &errorsutil.AppError{Code: errorsutil.ErrorNotFound, Err: err, Message: errorUserNotFound}
	}

	return nil
}

func (r *repository) GetUserByEmail(ctx *context.AppContext, email string) (*models.User, error) {
	var user models.User
	err := r.dbPool.QueryRow(ctx.StdContext(), r.queries.GetUserByEmail(), email).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errorsutil.SQLErrorParser(err)
	}

	return &user, nil
}

func (r *repository) GetUserByUsername(ctx *context.AppContext, username string) (*models.User, error) {
	var user models.User
	err := r.dbPool.QueryRow(ctx.StdContext(), r.queries.GetUserByUsername(), username).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errorsutil.SQLErrorParser(err)
	}

	return &user, nil
}

func (r *repository) GetAllUsers(ctx *context.AppContext) ([]*models.User, error) {
	rows, err := r.dbPool.Query(ctx.StdContext(), r.queries.GetAllUsers())
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
