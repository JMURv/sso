package db

const roleSelect = `
SELECT COUNT(*) 
FROM roles r %s`

const roleList = `
SELECT 
	r.id,
	r.name,
	r.description,
	ARRAY_AGG(p.id || '|' || p.name || '|' || p.description) FILTER (WHERE p.id IS NOT NULL) AS permissions
FROM roles r
LEFT JOIN role_permissions rp ON rp.role_id = r.id
LEFT JOIN permission p ON p.id = rp.permission_id
%s
GROUP BY r.id
ORDER BY name 
LIMIT $%d OFFSET $%d
`
const roleGet = `SELECT id, name, description FROM roles WHERE id = $1`
const roleCreate = `INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id`
const roleUpdate = `UPDATE roles SET name = $1, description = $2 WHERE id = $3`
const roleDelete = `DELETE FROM roles WHERE id = $1`

const roleAddPermQ = `
INSERT INTO role_permissions (role_id, permission_id) 
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING
`

const roleRemovePermQ = `
DELETE FROM role_permissions 
WHERE role_id = $1
`
