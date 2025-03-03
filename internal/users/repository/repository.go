package repository

import (
	"MydroX/project-v/internal/common/errorscode"
	"MydroX/project-v/internal/users/models"
	"MydroX/project-v/pkg/logger"
	"context"
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

// NewRepository is creating an interface for every method of the repository
func NewRepository(l *logger.Logger, dbPool *pgxpool.Pool) UsersRepository {
	return &repository{
		logger: l,
		dbPool: dbPool,
	}
}

func (r *repository) CreateUser(ctx *context.Context, user *models.User) error {
	query := `INSERT INTO users (uuid, username, email, password, role) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.dbPool.Exec(*ctx, query, user.UUID, user.Username, user.Email, user.Password, user.Role)

	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)

		if pgErr.Code == pgerrcode.UniqueViolation {
			*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_DUPLICATE_ENTITY)
			return errorUserAlreadyExists
		}

		r.logger.Zap.Sugar().Errorf("error creating user: %v", err)
		return fmt.Errorf("error creating user: %v", err)
	}

	return nil
}

func (r *repository) GetUserByUUID(ctx *context.Context, uuid string) (*models.User, error) {
	query := `SELECT uuid, username, email, password, role, created_at, updated_at FROM users WHERE uuid = $1`

	var user models.User
	err := r.dbPool.QueryRow(*ctx, query, uuid).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
			return nil, errorUserNotFound
		}

		r.logger.Zap.Sugar().Errorf("error getting user by uuid: %v", err)
		return nil, fmt.Errorf("error getting user by uuid: %v", err)
	}

	return &user, nil
}

func (r *repository) UpdateUser(ctx *context.Context, user *models.User) (*models.User, error) {
	query := `UPDATE users SET username = $1, email = $2, role = $3 WHERE uuid = $4 RETURNING uuid, username, email, password, role, created_at, updated_at`

	err := r.dbPool.QueryRow(*ctx, query, user.Username, user.Email, user.Role, user.UUID).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
			return nil, errorUserNotFound
		}
		r.logger.Zap.Sugar().Errorf("error updating user: %v", err)
		return nil, fmt.Errorf("error updating user: %v", err)
	}
	return user, nil
}

func (r *repository) UpdatePassword(ctx *context.Context, uuid, password string) error {
	query := `UPDATE users SET password = $1 WHERE uuid = $2`

	res, err := r.dbPool.Exec(*ctx, query, password, uuid)
	if err != nil {
		r.logger.Zap.Sugar().Errorf("error updating password: %v", err)
		return fmt.Errorf("error updating password: %v", err)
	}

	if res.RowsAffected() == 0 {
		*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
		return errorUserNotFound
	}

	return nil
}

func (r *repository) UpdateEmail(ctx *context.Context, uuid, email string) error {
	query := `UPDATE users SET email = $1 WHERE uuid = $2`

	res, err := r.dbPool.Exec(*ctx, query, email, uuid)
	if err != nil {
		r.logger.Zap.Sugar().Errorf("error updating email: %v", err)
		return fmt.Errorf("error updating email: %v", err)
	}

	if res.RowsAffected() == 0 {
		*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
		return errorUserNotFound
	}

	return nil
}

func (r *repository) DeleteUser(ctx *context.Context, uuid string) error {
	query := `DELETE FROM users WHERE uuid = $1`

	res, err := r.dbPool.Exec(*ctx, query, uuid)
	if err != nil {
		r.logger.Zap.Sugar().Errorf("error deleting user: %v", err)
		return fmt.Errorf("error deleting user: %v", err)
	}

	if res.RowsAffected() == 0 {
		*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
		return errorUserNotFound
	}

	return nil
}

func (r *repository) GetUserByEmail(ctx *context.Context, email string) (*models.User, error) {
	query := `SELECT uuid, username, email, password, role, created_at, updated_at FROM users WHERE email = $1`

	var user models.User
	err := r.dbPool.QueryRow(*ctx, query, email).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
			return nil, errorUserNotFound
		}
		r.logger.Zap.Sugar().Errorf("error getting user by email: %v", err)
		return nil, fmt.Errorf("error getting user by email: %v", err)
	}

	return &user, nil
}

func (r *repository) GetUserByUsername(ctx *context.Context, username string) (*models.User, error) {
	query := `SELECT uuid, username, email, password, role, created_at, updated_at FROM users WHERE username = $1`

	var user models.User
	err := r.dbPool.QueryRow(*ctx, query, username).Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
			return nil, errorUserNotFound
		}
		r.logger.Zap.Sugar().Errorf("error getting user by username: %v", err)
		return nil, fmt.Errorf("error getting user by username: %v", err)
	}

	return &user, nil
}

func (r *repository) GetAllUsers(ctx *context.Context) ([]*models.User, error) {
	query := `SELECT uuid, username, email, password, role, created_at, updated_at FROM users`

	rows, err := r.dbPool.Query(*ctx, query)
	if err != nil {
		r.logger.Zap.Sugar().Errorf("error getting all users: %v", err)
		return nil, fmt.Errorf("error getting all users: %v", err)
	}

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			r.logger.Zap.Sugar().Errorf("error scanning user: %v", err)
			return nil, fmt.Errorf("error scanning user: %v", err)
		}
		users = append(users, &user)
	}

	return users, nil
}
