//revive:disable:line-length-limit
package repository

type Queries struct{}

func (_ *Queries) SaveSession() string {
	return `INSERT INTO sessions
	(uuid, user_uuid, refresh_token_hash, last_used_at, os, os_version, browser, browser_version, ipv4_address, created_at, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
}

func (_ *Queries) GetAllowedAudiences() string {
	return `SELECT audience FROM allowed_audiences WHERE active = true`
}

func (_ *Queries) GetUserAudiences() string {
	return `SELECT aa.audience FROM allowed_audiences aa JOIN user_audiences ua ON aa.uuid = ua.audience_uuid WHERE ua.user_uuid = $1 AND aa.active = true`
}

func (_ *Queries) IsValidAudience() string {
	return `SELECT EXISTS(SELECT 1 FROM allowed_audiences WHERE audience = $1 AND active = true)`
}

func (_ *Queries) RegisterAudience() string {
	return `INSERT INTO allowed_audiences (uuid, audience, service_name, description, permissions) VALUES ($1, $2, $3, $4, $5)`
}

func (_ *Queries) RevokeAudience() string {
	return `UPDATE allowed_audiences SET active = false, updated_at = NOW() WHERE audience = $1`
}

func (_ *Queries) AssignAudienceToUser() string {
	return `INSERT INTO user_audiences (user_uuid, audience_uuid) SELECT $1, uuid FROM allowed_audiences WHERE audience = $2 AND active = true`
}

func (_ *Queries) UnassignAudienceFromUser() string {
	return `DELETE FROM user_audiences WHERE user_uuid = $1 AND audience_uuid = (SELECT uuid FROM allowed_audiences WHERE audience = $2)`
}
