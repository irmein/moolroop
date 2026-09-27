package model

// ErrorResponse represents a standardized error envelope returned by HTTP endpoints.
type ErrorResponse struct {
	Error string `json:"error" example:"user profile not found"`
}
