package model

import (
	"strings"

	"moolroop/internal/errors"
)

// LogActivityPayload represents the payload to log an immutable activity.
type LogActivityPayload struct {
	UserID      string `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	ActionType  string `json:"action_type" binding:"required" example:"PROFILE_UPDATE"`
	Description string `json:"description" example:"User updated profile contact info"`
}

// Validate validates the LogActivityPayload.
func (p *LogActivityPayload) Validate() error {
	trimmedUserID := strings.TrimSpace(p.UserID)
	if trimmedUserID == "" {
		return errors.ErrEmptyUserID
	}

	trimmedAction := strings.TrimSpace(p.ActionType)
	if trimmedAction == "" {
		return errors.ErrEmptyActionType
	}

	p.UserID = trimmedUserID
	p.ActionType = trimmedAction
	p.Description = strings.TrimSpace(p.Description)
	return nil
}
