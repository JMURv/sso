package db

const userSelectQ = `
SELECT COUNT(*)
FROM users
`

const userSearchSelectQ = `
SELECT COUNT(*)
FROM users 
WHERE name ILIKE $1 OR email ILIKE $2
`

const userSearchQ = `
SELECT 
	u.id, 
	u.name, 
	u.email, 
	u.avatar, 
	u.created_at, 
	u.updated_at,
	ARRAY_AGG(r.id || '|' || r.name || '|' || r.description) FILTER (WHERE r.id IS NOT NULL) AS roles
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
WHERE u.name ILIKE $1 OR u.email ILIKE $2
GROUP BY u.id, u.name
ORDER BY u.name DESC 
LIMIT $3 OFFSET $4
`

const userListQ = `
SELECT 
	u.id, 
	u.name, 
	u.email, 
	u.avatar, 
	u.created_at, 
	u.updated_at,
	ARRAY_AGG(r.id || '|' || r.name || '|' || r.description) FILTER (WHERE r.id IS NOT NULL) AS roles
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
GROUP BY u.id, u.created_at
ORDER BY created_at DESC 
LIMIT $1 OFFSET $2
`

const userGetByIDQ = `
SELECT 
	u.id, 
	u.name, 
	u.email, 
	u.avatar,
	u.is_wa,
	u.is_active,
	u.is_email_verified,
	u.created_at, 
	u.updated_at,
	ARRAY_AGG(r.id || '|' || r.name || '|' || r.description) FILTER (WHERE r.id IS NOT NULL) AS roles,
	ARRAY_AGG(oth2.provider || '|' || oth2.provider_id) FILTER (WHERE oth2.id IS NOT NULL) AS oauth2_connections	
FROM users u
LEFT JOIN oauth2_connections oth2 ON oth2.user_id = u.id
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
WHERE u.id = $1
GROUP BY u.id
`

const userGetByEmailQ = `
SELECT 
    u.id, 
    u.name, 
    u.email, 
    u.password,
    u.avatar,
	u.is_wa,
	u.is_active,
	u.is_email_verified,
    u.created_at, 
    u.updated_at,
    ARRAY_AGG(r.id || '|' || r.name || '|' || r.description) FILTER (WHERE r.id IS NOT NULL) AS roles
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
WHERE email = $1
GROUP BY u.id
`

const userCreateQ = `
INSERT INTO users (name, password, email, avatar, is_active, is_email_verified) 
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id
`

const userUpdateQ = `
UPDATE users 
SET name = $1, 
    password = $2,
    email = $3,
    avatar = $4,
	is_active = $5,
	is_email_verified = $6
WHERE id = $7`

const userDeleteQ = `
DELETE FROM users 
WHERE id = $1
`

const userAddRoleQ = `
INSERT INTO user_roles (user_id, role_id) 
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING
`

const userRemoveRoleQ = `
DELETE FROM user_roles 
WHERE user_id = $1
`
