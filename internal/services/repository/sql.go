//revive:disable:line-length-limit
package repository

type ServiceQueries struct{}

func (_ *ServiceQueries) GetAllowedServices() string {
	return `SELECT audience FROM allowed_audiences WHERE active = true`
}

func (_ *ServiceQueries) GetUserServices() string {
	return `SELECT aa.audience FROM allowed_audiences aa JOIN user_audiences ua ON aa.uuid = ua.audience_uuid WHERE ua.user_uuid = $1 AND aa.active = true`
}

func (_ *ServiceQueries) IsValidService() string {
	return `SELECT EXISTS(SELECT 1 FROM allowed_audiences WHERE audience = $1 AND active = true)`
}

func (_ *ServiceQueries) RegisterService() string {
	return `INSERT INTO allowed_audiences (uuid, audience, service_name, description, permissions) VALUES ($1, $2, $3, $4, $5)`
}

func (_ *ServiceQueries) RevokeService() string {
	return `UPDATE allowed_audiences SET active = false, updated_at = NOW() WHERE audience = $1`
}

func (_ *ServiceQueries) AssignServiceToUser() string {
	return `INSERT INTO user_audiences (user_uuid, audience_uuid) SELECT $1, uuid FROM allowed_audiences WHERE audience = $2 AND active = true`
}

func (_ *ServiceQueries) UnassignServiceFromUser() string {
	return `DELETE FROM user_audiences WHERE user_uuid = $1 AND audience_uuid = (SELECT uuid FROM allowed_audiences WHERE audience = $2)`
}
