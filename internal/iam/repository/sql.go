//revive:disable:unused-receiver
package repository

type Queries struct{}

func (q *Queries) SaveSession() string {
	return `INSERT INTO sessions 
	(uuid, user_uuid, refresh_token, last_used_at, os, browser, browser_version, ipv4_addres, created_at, expires_at, ) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
}
