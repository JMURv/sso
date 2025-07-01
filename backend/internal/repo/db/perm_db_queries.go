package db

const (
	permSelect = `SELECT COUNT(*) FROM permission p %s`
	permList   = `
SELECT
	p.id,
	p.name,
	p.description
FROM permission p
%s
ORDER BY name
LIMIT $%d OFFSET $%d
`
)

const (
	permGet    = `SELECT id, name, description FROM permission WHERE id = $1`
	permCreate = `INSERT INTO permission (name, description) VALUES ($1, $2) RETURNING id`
	permUpdate = `UPDATE permission SET name = $1, description = $2 WHERE id = $3`
	permDelete = `DELETE FROM permission WHERE id = $1`
)
