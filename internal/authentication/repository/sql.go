//revive:disable:line-length-limit
package repository

type SessionQueries struct{}

func (_ *SessionQueries) SaveSession() string {
	return `INSERT INTO sessions
	(uuid, user_uuid, refresh_token_hash, last_used_at, os, os_version, browser, browser_version, ipv4_address, created_at, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
}

func (_ *SessionQueries) GetSessionByUUID() string {
	return `SELECT uuid, user_uuid, refresh_token_hash, last_used_at, os, os_version, browser, browser_version, ipv4_address, created_at, expires_at FROM sessions WHERE uuid = $1`
}

func (_ *SessionQueries) DeleteSession() string {
	return `DELETE FROM sessions WHERE uuid = $1`
}

func (_ *SessionQueries) DeleteAllUserSessions() string {
	return `DELETE FROM sessions WHERE user_uuid = $1`
}

func (_ *SessionQueries) UpdateSessionLastUsed() string {
	return `UPDATE sessions SET last_used_at = NOW() WHERE uuid = $1`
}
