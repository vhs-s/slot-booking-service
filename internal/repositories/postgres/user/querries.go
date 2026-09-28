package user

const CREATE string = "INSERT INTO users (id, name, email, password_hash, created_at, updated_at) values ($1,$2,$3,$4,$5,$6);"
const GET_BY_ID string = "SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1;"
const GET_BY_EMAIL string = "SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = $1;"
