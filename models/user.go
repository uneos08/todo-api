package models

type User struct {
	ID           int    `db:"id" json:"id"`
	Username     string `db:"username" json:"username"`
	PasswordHash string `db:"password_hash" json:"-"`
}
