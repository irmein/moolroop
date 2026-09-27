package model

import (
	"net/mail"
	"strings"

	"github.com/moolroop/internal/errors"
)

// CreateProfilePayload represents the payload to create a new profile.
type CreateProfilePayload struct {
	FullName string `json:"full_name" binding:"required" example:"Alice Smith"`
	Email    string `json:"email" binding:"required,email" example:"alice@example.com"`
}

// Validate validates the CreateProfilePayload.
func (p *CreateProfilePayload) Validate() error {
	trimmedName := strings.TrimSpace(p.FullName)
	if trimmedName == "" {
		return errors.ErrEmptyName
	}

	trimmedEmail := strings.TrimSpace(p.Email)
	if trimmedEmail == "" {
		return errors.ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(trimmedEmail)
	if err != nil || addr.Address != trimmedEmail {
		return errors.ErrInvalidEmail
	}

	p.FullName = trimmedName
	p.Email = strings.ToLower(trimmedEmail)
	return nil
}

// PatchProfilePayload represents the fields that can be updated in a profile.
type PatchProfilePayload struct {
	FullName *string `json:"full_name,omitempty" example:"Alice Cooper"`
	Email    *string `json:"email,omitempty" example:"alice.new@example.com"`
}

// Validate validates the PatchProfilePayload.
func (p *PatchProfilePayload) Validate() error {
	if p.FullName == nil && p.Email == nil {
		return errors.ErrInvalidPayload
	}

	if p.FullName != nil {
		trimmedName := strings.TrimSpace(*p.FullName)
		if trimmedName == "" {
			return errors.ErrEmptyName
		}
		*p.FullName = trimmedName
	}

	if p.Email != nil {
		trimmedEmail := strings.TrimSpace(*p.Email)
		addr, err := mail.ParseAddress(trimmedEmail)
		if err != nil || addr.Address != trimmedEmail {
			return errors.ErrInvalidEmail
		}
		*p.Email = strings.ToLower(trimmedEmail)
	}

	return nil
}
