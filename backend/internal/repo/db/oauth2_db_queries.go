package db

const getUserOauth2 = `
SELECT 
    u.id,
    u.email,
    u.name,
    u.avatar,
    ARRAY_AGG(r.id || '|' || r.name || '|' || r.description) FILTER (WHERE r.id IS NOT NULL) AS roles
FROM users u
JOIN oauth2_connections oc ON u.id = oc.user_id
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
WHERE oc.provider = $1 AND oc.provider_id = $2
GROUP BY u.id
`

const createOAuth2Connection = `
INSERT INTO oauth2_connections 
(user_id, provider, provider_id, access_token, refresh_token, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
`
