package models

type Todo struct {
	ID       int     `json:"id,omitempty" swaggerignore:"true"`
	Title    string  `json:"title"`
	Done     bool    `json:"done"`
	UserID   int     `json:"user_id,omitempty" db:"user_id" swaggerignore:"true"`
	PhotoURL *string `json:"photo_url,omitempty" db:"photo_url"`
}
