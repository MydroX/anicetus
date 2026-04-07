//revive:disable:line-length-limit
package repository

type IdentityQueries struct{}

func (_ *IdentityQueries) CreateUser() string {
	return `INSERT INTO users (uuid, username, email, password) VALUES ($1, $2, $3, $4)`
}

func (_ *IdentityQueries) GetUserByUUID() string {
	return `SELECT uuid, username, email, password, created_at, updated_at FROM users WHERE uuid = $1`
}

func (_ *IdentityQueries) UpdateUser() string {
	return `UPDATE users SET username = $1, email = $2 WHERE uuid = $3 RETURNING uuid, username, email, password, created_at, updated_at`
}

func (_ *IdentityQueries) UpdatePassword() string {
	return `UPDATE users SET password = $1 WHERE uuid = $2`
}

func (_ *IdentityQueries) UpdateEmail() string {
	return `UPDATE users SET email = $1 WHERE uuid = $2`
}

func (_ *IdentityQueries) DeleteUser() string {
	return `DELETE FROM users WHERE uuid = $1`
}

func (_ *IdentityQueries) GetUserByUsername() string {
	return `SELECT uuid, username, email, password, created_at, updated_at FROM users WHERE username = $1`
}

func (_ *IdentityQueries) GetUserByEmail() string {
	return `SELECT uuid, username, email, password, created_at, updated_at FROM users WHERE email = $1`
}

func (_ *IdentityQueries) GetAllUsers() string {
	return `SELECT uuid, username, email, created_at, updated_at FROM users`
}

func (_ *IdentityQueries) GetAllUsersOffset() string {
	return `SELECT uuid, username, email, created_at, updated_at FROM users OFFSET $1 LIMIT $2`
}
