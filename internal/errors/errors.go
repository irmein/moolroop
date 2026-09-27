package errors

import (
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserNotFound    = errors.New("user profile not found")
	ErrInvalidEmail   = errors.New("invalid email provided")
	ErrEmptyName       = errors.New("full name cannot be empty")
	ErrEmptyUserID     = errors.New("user_id cannot be empty")
	ErrEmptyActionType = errors.New("action_type cannot be empty")
	ErrInvalidPayload  = errors.New("invalid request payload")
)

// ToGRPCStatus maps domain errors to appropriate gRPC status errors.
func ToGRPCStatus(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidEmail),
		errors.Is(err, ErrEmptyName),
		errors.Is(err, ErrEmptyUserID),
		errors.Is(err, ErrEmptyActionType),
		errors.Is(err, ErrInvalidPayload):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// ToHTTPStatus maps domain errors to HTTP status codes.
func ToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch {
	case errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInvalidEmail),
		errors.Is(err, ErrEmptyName),
		errors.Is(err, ErrEmptyUserID),
		errors.Is(err, ErrEmptyActionType),
		errors.Is(err, ErrInvalidPayload):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
