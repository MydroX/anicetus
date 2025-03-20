package repository

import (
	"MydroX/anicetus/internal/common/context"
	errorsutil "MydroX/anicetus/internal/common/errors"
	"MydroX/anicetus/internal/users/models"
	"MydroX/anicetus/pkg/logger"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	errorUserAlreadyExists = errors.New("user already exists")
	errorUserNotFound      = errors.New("user not found")
)

type repository struct {
	logger *logger.Logger
	dbPool *pgxpool.Pool
}

// New is creating an interface for every method of the repository
func New(l *logger.Logger, dbPool *pgxpool.Pool) UsersRepository {
	return &repository{
		logger: l,
		dbPool: dbPool,
	}
}

func (r *repository) CreateUser(ctx *context.AppContext, user *models.User) *errorsutil.Err {
	query := `INSERT INTO users (uuid, username, email, password, role) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.dbPool.Exec(ctx.StdContext(), query, user.UUID, user.Username, user.Email, user.Password, user.Role)

	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)

		if pgErr.Code == pgerrcode.UniqueViolation {
			return &errorsutil.Err{Code: errorsutil.ERROR_DUPLICATE_ENTITY, Err: errorUserAlreadyExists}
		}

		r.logger.Zap.Sugar().Errorf("error creating user: %v", err)
		return &errorsutil.Err{
			Code: errorsutil.ERROR_INTERNAL,
			Err:  fmt.Errorf("error creating user: %v", err),
		}
	}

	return nil
}

func (r *repository) GetUserByUUID(ctx *context.AppContext, uuid string) (*models.User, *errorsutil.Err) {
	query := `SELECT uuid, username, email, password, role, created_at, updated_at FROM users WHERE uuid = $1`

	var user models.User
	err := r.dbPool.QueryRow(ctx.StdContext(), query, uuid).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &errorsutil.Err{Code: errorsutil.ERROR_NOT_FOUND, Err: errorUserNotFound}
		}

		r.logger.Zap.Sugar().Errorf("error getting user by uuid: %v", err)
		return nil, &errorsutil.Err{
			Code: errorsutil.ERROR_INTERNAL,
			Err:  fmt.Errorf("error getting user by uuid: %v", err),
		}
	}

	return &user, nil
}

func (r *repository) UpdateUser(ctx *context.AppContext, user *models.User) (*models.User, *errorsutil.Err) {
	query := `UPDATE users SET username = $1, email = $2, role = $3 WHERE uuid = $4 RETURNING uuid, username, email, password, role, created_at, updated_at`

	err := r.dbPool.QueryRow(ctx.StdContext(), query, user.Username, user.Email, user.Role, user.UUID).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &errorsutil.Err{Code: errorsutil.ERROR_NOT_FOUND, Err: errorUserNotFound}
		}
		r.logger.Zap.Sugar().Errorf("error updating user: %v", err)
		return nil, &errorsutil.Err{
			Code: errorsutil.ERROR_INTERNAL,
			Err:  fmt.Errorf("error updating user: %v", err),
		}
	}
	return user, nil
}

func (r *repository) UpdatePassword(ctx *context.AppContext, uuid, password string) *errorsutil.Err {
	query := `UPDATE users SET password = $1 WHERE uuid = $2`

	res, err := r.dbPool.Exec(ctx.StdContext(), query, password, uuid)
	if err != nil {
		r.logger.Zap.Sugar().Errorf("error updating password: %v", err)
		return &errorsutil.Err{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("error updating password: %v", err)}
	}

	if res.RowsAffected() == 0 {
		return &errorsutil.Err{Code: errorsutil.ERROR_NOT_FOUND, Err: errorUserNotFound}
	}

	return nil
}

func (r *repository) UpdateEmail(ctx *context.AppContext, uuid, email string) *errorsutil.Err {
	query := `UPDATE users SET email = $1 WHERE uuid = $2`

	res, err := r.dbPool.Exec(ctx.StdContext(), query, email, uuid)
	if err != nil {
		r.logger.Zap.Sugar().Errorf("error updating email: %v", err)
		return &errorsutil.Err{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("error updating email: %v", err)}
	}

	if res.RowsAffected() == 0 {
		return &errorsutil.Err{Code: errorsutil.ERROR_NOT_FOUND, Err: errorUserNotFound}
	}

	return nil
}

func (r *repository) DeleteUser(ctx *context.AppContext, uuid string) *errorsutil.Err {
	query := `DELETE FROM users WHERE uuid = $1`

	res, err := r.dbPool.Exec(ctx.StdContext(), query, uuid)
	if err != nil {
		r.logger.Zap.Sugar().Errorf("error deleting user: %v", err)
		return &errorsutil.Err{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("error deleting user: %v", err)}
	}

	if res.RowsAffected() == 0 {
		return &errorsutil.Err{Code: errorsutil.ERROR_NOT_FOUND, Err: errorUserNotFound}
	}

	return nil
}

func (r *repository) GetUserByEmail(ctx *context.AppContext, email string) (*models.User, *errorsutil.Err) {
	query := `SELECT uuid, username, email, password, role, created_at, updated_at FROM users WHERE email = $1`

	var user models.User
	err := r.dbPool.QueryRow(ctx.StdContext(), query, email).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &errorsutil.Err{Code: errorsutil.ERROR_NOT_FOUND, Err: errorUserNotFound}
		}
		r.logger.Zap.Sugar().Errorf("error getting user by email: %v", err)
		return nil, &errorsutil.Err{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("error getting user by email: %v", err)}
	}

	return &user, nil
}

func (r *repository) GetUserByUsername(ctx *context.AppContext, username string) (*models.User, *errorsutil.Err) {
	query := `SELECT uuid, username, email, password, role, created_at, updated_at FROM users WHERE username = $1`

	var user models.User
	err := r.dbPool.QueryRow(ctx.StdContext(), query, username).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &errorsutil.Err{Code: errorsutil.ERROR_NOT_FOUND, Err: errorUserNotFound}
		}
		r.logger.Zap.Sugar().Errorf("error getting user by username: %v", err)
		return nil, &errorsutil.Err{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("error getting user by username: %v", err)}
	}

	return &user, nil
}

func (r *repository) GetAllUsers(ctx *context.AppContext) ([]*models.User, *errorsutil.Err) {
	query := `SELECT uuid, username, email, password, role, created_at, updated_at FROM users`

	rows, err := r.dbPool.Query(ctx.StdContext(), query)
	if err != nil {
		r.logger.Zap.Sugar().Errorf("error getting all users: %v", err)
		return nil, &errorsutil.Err{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("error getting all users: %v", err)}
	}

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			r.logger.Zap.Sugar().Errorf("error scanning user: %v", err)
			return nil, &errorsutil.Err{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("error scanning user: %v", err)}
		}
		users = append(users, &user)
	}

	return users, nil
}
