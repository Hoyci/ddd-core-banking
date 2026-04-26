package http

type ApiResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}
