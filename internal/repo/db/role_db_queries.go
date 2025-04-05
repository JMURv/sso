package db

const roleSelect = `SELECT COUNT(*) FROM roles`
const roleList = `SELECT id, name, description FROM roles ORDER BY name LIMIT $1 OFFSET $2`
const roleGet = `SELECT id, name, description FROM roles WHERE id = $1`
const roleCreate = `INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id`
const roleUpdate = `UPDATE roles SET name = $1, description = $2 WHERE id = $3`
const roleDelete = `DELETE FROM roles WHERE id = $1`
