package models

type GeneralResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"` //  []string или map[string]string
	Meta    any    `json:"meta,omitempty"`   // Доп. инфо (пагинация и др.)
}
