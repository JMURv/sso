package db

const permSelect = `SELECT COUNT(*) FROM permission`
const permList = `SELECT id, name, description FROM permission ORDER BY name LIMIT $1 OFFSET $2`
const permGet = `SELECT id, name, description FROM permission WHERE id = $1`
const permCreate = `INSERT INTO permission (name, description) VALUES ($1, $2) RETURNING id`
const permUpdate = `UPDATE permission SET name = $1, description = $2 WHERE id = $3`
const permDelete = `DELETE FROM permission WHERE id = $1`
