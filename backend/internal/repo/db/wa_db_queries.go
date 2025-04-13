package db

const getWACredentials = `
SELECT 
	id, 
	public_key,
	attestation_type,
	authenticator
FROM wa_credentials
WHERE user_id = $1
`

const createWACredential = `
INSERT INTO wa_credentials 
(id, public_key, attestation_type, authenticator, user_id)
VALUES ($1, $2, $3, $4, $5)
`

const setIsWA = `
UPDATE users
SET is_wa = TRUE
WHERE id = $1
`

const updateWACredentials = `
UPDATE wa_credentials
SET
	public_key = $1,
	attestation_type = $2,
	authenticator = $3
WHERE id = $4
`

const deleteWACredential = `
DELETE FROM wa_credentials
WHERE id = $1 AND user_id = $2
`
