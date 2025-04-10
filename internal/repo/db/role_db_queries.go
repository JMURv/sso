package db

const roleSelect = `SELECT COUNT(*) FROM roles`
const roleList = `
SELECT 
	r.id,
	r.name,
	r.description,
	ARRAY_AGG(p.id || '|' || p.name || '|' || p.description) FILTER (WHERE p.id IS NOT NULL) AS permissions
FROM roles r
LEFT JOIN role_permissions rp ON rp.role_id = r.id
LEFT JOIN permission p ON p.id = rp.permission_id
GROUP BY r.id
ORDER BY name 
LIMIT $1 OFFSET $2
`
const roleGet = `SELECT id, name, description FROM roles WHERE id = $1`
const roleCreate = `INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id`
const roleUpdate = `UPDATE roles SET name = $1, description = $2 WHERE id = $3`
const roleDelete = `DELETE FROM roles WHERE id = $1`

const roleSearchSelectQ = `
SELECT COUNT(*)
FROM roles 
WHERE name ILIKE $1
`

const roleSearchQ = `
SELECT 
	ur.id, 
	ur.name, 
	ur.description 
FROM roles ur
WHERE ur.name ILIKE $1
LIMIT $2 OFFSET $3
`
