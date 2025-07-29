package models

type Todo struct {
	ID    int    `json:"id,omitempty" swaggerignore:"true"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}
