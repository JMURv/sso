package db

const userSelectQ = `
SELECT COUNT(DISTINCT u.id)
FROM users u %s
`

const userListQ = `
SELECT 
	u.id, 
	u.name, 
	u.email, 
	u.avatar, 
	u.created_at, 
	u.updated_at,
	ARRAY_AGG(r.id || '|' || r.name || '|' || r.description) FILTER (WHERE r.id IS NOT NULL) AS roles,
	ARRAY_AGG(oth2.provider || '|' || oth2.provider_id) FILTER (WHERE oth2.id IS NOT NULL) AS oauth2_connections,
	ARRAY_AGG(
		DISTINCT ud.id || '|' || 
		ud.name || '|' || 
		ud.device_type || '|' ||
		ud.os || '|' || 
		ud.user_agent || '|' ||
		ud.browser || '|' || 
		ud.ip || '|' ||
		ud.last_active
	) FILTER (WHERE ud.id IS NOT NULL) AS devices
FROM users u
LEFT JOIN user_devices ud ON ud.user_id = u.id
LEFT JOIN oauth2_connections oth2 ON oth2.user_id = u.id
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
%s
GROUP BY u.id
ORDER BY %s 
LIMIT $%d OFFSET $%d
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
