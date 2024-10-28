package db

const permSelect = `SELECT COUNT(*) FROM permission`
const permList = `SELECT id, name FROM permission ORDER BY name LIMIT $1 OFFSET $2`
const permGet = `SELECT id, name FROM permission WHERE id = $1`
const permCreate = `INSERT INTO permission (name) VALUES ($1) RETURNING id`
const permUpdate = `UPDATE permission SET name = $1 WHERE id = $2`
const permDelete = `DELETE FROM permission WHERE id = $1`
