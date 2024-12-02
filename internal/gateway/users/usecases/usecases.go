// Package usecases is the internal logic. Describes an action that the user wants to perform.
// Also interact with repository and determines how the data has to be transmitted to the external layer.
package usecases

import (
	"MydroX/project-v/internal/gateway/users/config"
	"MydroX/project-v/internal/gateway/users/models"
	"MydroX/project-v/internal/gateway/users/repository"
	"MydroX/project-v/pkg/jwt"
	"MydroX/project-v/pkg/logger"
	"MydroX/project-v/pkg/password"
	passwordPkg "MydroX/project-v/pkg/password"
	"MydroX/project-v/pkg/uuid"
	"fmt"
	"time"
)

var prefix = "user"

type usecases struct {
	logger     *logger.Logger
	repository repository.UsersRepository
	jwtConfig  *config.JWT
}

// NewUsecases is creating an interface for all the usecases of the service.
func NewUsecases(l *logger.Logger, r repository.UsersRepository, jwtConfig *config.JWT) UsersUsecases {
	return &usecases{
		logger:     l,
		repository: r,
		jwtConfig:  jwtConfig,
	}
}

func (u *usecases) Create(req *models.User) error {
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
	}

	user.Password, _ = passwordPkg.Hash(req.Password)
	user.UUID = uuid.NewWithPrefix(prefix)

	err := u.repository.CreateUser(&user)

	return err
}

func (u *usecases) Get(uuid string) (*models.User, error) {
	user, err := u.repository.GetUser(uuid)
	if err != nil {
		return nil, err
	}

	res := models.User{
		UUID:     user.UUID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	return &res, err
}

func (u *usecases) Update(user *models.User) error {
	err := u.repository.UpdateUser(user)
	return err
}

func (u *usecases) UpdatePassword(uuid string, newPassword string) error {
	newPasswordCrypted, _ := password.Hash(newPassword)

	err := u.repository.UpdatePassword(uuid, newPasswordCrypted)
	return err
}

func (u *usecases) UpdateEmail(uuid string, email string) error {
	err := u.repository.UpdateEmail(uuid, email)
	return err
}

func (u *usecases) Delete(uuid string) error {
	err := u.repository.DeleteUser(uuid)
	return err
}

func (u *usecases) Login(username, email, password string) (string, error) {
	switch {
	case email != "":
		user, err := u.repository.GetUserByEmail(email)
		if err != nil {
			return "", err
		}

		token, err := login(u.jwtConfig, user.UUID, password, user.Password)

		return token, err

	case username != "":
		user, err := u.repository.GetUserByUsername(username)
		if err != nil {
			return "", err
		}

		token, err := login(u.jwtConfig, user.UUID, password, user.Password)

		return token, err
	}
	return "", fmt.Errorf("username or email must be provided")
}

func login(jwtConfig *config.JWT, userUUID, password, passwordCrypted string) (string, error) {
	if !passwordPkg.CheckPasswordHash(password, passwordCrypted) {
		return "", fmt.Errorf("invalid password")
	}
	expirationTime := time.Now().Add(time.Second * time.Duration(jwtConfig.ExpirationTime))

	token, err := jwt.CreateToken(
		expirationTime,
		jwtConfig.Secret,
		userUUID,
	)
	if err != nil {
		return "", fmt.Errorf("error creating token: %v", err)
	}

	return token, nil
}
