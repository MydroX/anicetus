//revive:disable:line-length-limit
package repository

type UsersQueries struct{}

func (_ *UsersQueries) CreateUser() string {
	return `INSERT INTO users (uuid, username, email, password, role) VALUES ($1, $2, $3, $4, $5)`
}

func (_ *UsersQueries) GetUserByUUID() string {
	return `SELECT uuid, username, email, password, role, created_at, updated_at FROM users WHERE uuid = $1`
}

func (_ *UsersQueries) UpdateUser() string {
	return `UPDATE users SET username = $1, email = $2, role = $3 WHERE uuid = $4 RETURNING uuid, username, email, password, role, created_at, updated_at`
}

func (_ *UsersQueries) UpdatePassword() string {
	return `UPDATE users SET password = $1 WHERE uuid = $2`
}

func (_ *UsersQueries) UpdateEmail() string {
	return `UPDATE users SET email = $1 WHERE uuid = $2`
}

func (_ *UsersQueries) UpdateRole() string {
	return `UPDATE users SET role = $1 WHERE uuid = $2`
}

func (_ *UsersQueries) DeleteUser() string {
	return `DELETE FROM users WHERE uuid = $1`
}

func (_ *UsersQueries) GetUserByUsername() string {
	return `SELECT uuid, username, email, password, role, created_at, updated_at FROM users WHERE username = $1`
}

func (_ *UsersQueries) GetUserByEmail() string {
	return `SELECT uuid, username, email, password, role, created_at, updated_at FROM users WHERE email = $1`
}

func (_ *UsersQueries) GetAllUsers() string {
	return `SELECT uuid, username, email, role, created_at, updated_at FROM users`
}

func (_ *UsersQueries) GetAllUsersOffset() string {
	return `SELECT uuid, username, email, role, created_at, updated_at FROM users OFFSET $1 LIMIT $2`
}
