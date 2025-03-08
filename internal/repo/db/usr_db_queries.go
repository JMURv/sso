package db

const userSelectQ = `SELECT COUNT(*) FROM users`

const userSearchSelectQ = `SELECT COUNT(*) FROM users WHERE name ILIKE $1 OR email ILIKE $2`

const userSearchQ = `
SELECT 
	u.id, 
	u.name, 
	u.password, 
	u.email, 
	u.avatar, 
	u.created_at, 
	u.updated_at,
	ARRAY_AGG(p.id::TEXT || '|' || p.name || '|' || up.value::TEXT) FILTER (WHERE p.id IS NOT NULL) AS permissions
FROM users u
LEFT JOIN user_permission up ON up.user_id = u.id
LEFT JOIN permission p ON p.id = up.permission_id
WHERE u.name ILIKE $1 OR u.email ILIKE $2
GROUP BY u.id, u.name
ORDER BY u.name DESC 
LIMIT $3 OFFSET $4
`

const userListQ = `
SELECT 
	u.id, 
	u.name, 
	u.password, 
	u.email, 
	u.avatar, 
	u.created_at, 
	u.updated_at,
	ARRAY_AGG(p.id::TEXT || '|' || p.name || '|' || up.value::TEXT) FILTER (WHERE p.id IS NOT NULL) AS permissions
FROM users u
LEFT JOIN user_permission up ON up.user_id = u.id
LEFT JOIN permission p ON p.id = up.permission_id
GROUP BY u.id, u.created_at
ORDER BY created_at DESC 
LIMIT $1 OFFSET $2
`

const userGetByIDQ = `
SELECT 
	u.id, 
	u.name, 
	u.password, 
	u.email, 
	u.avatar, 
	u.created_at, 
	u.updated_at,
	ARRAY_AGG(p.id::TEXT || '|' || p.name || '|' || up.value::TEXT) FILTER (WHERE p.id IS NOT NULL) AS permissions
FROM users u
LEFT JOIN user_permission up ON up.user_id = u.id
LEFT JOIN permission p ON p.id = up.permission_id
WHERE u.id = $1
GROUP BY u.id, u.created_at
`

const userGetByEmailQ = `
SELECT id, name, password, email, avatar, created_at, updated_at
FROM users
WHERE email = $1`

const userCreateQ = `
INSERT INTO users (name, password, email, avatar) 
VALUES ($1, $2, $3, $4)
RETURNING id
`

const userUpdateQ = `
UPDATE users 
SET name = $1, password = $2, email = $3, avatar = $4
WHERE id = $5`

const userDeleteQ = `DELETE FROM users WHERE id = $1`
const userCreatePermQ = `INSERT INTO user_permission (user_id, permission_id, value) VALUES ($1, $2, $3)`
const userDeletePermQ = `DELETE FROM user_permission WHERE user_id = $1`
