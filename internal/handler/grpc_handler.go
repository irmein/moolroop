package handler

import (
	"context"
	"strings"

	v1 "github.com/moolroop/gen/v1"
	apperrors "github.com/moolroop/internal/errors"
	"github.com/moolroop/internal/model"
	"github.com/moolroop/internal/store"
)

type GrpcHandler struct {
	v1.UnimplementedIdentityServiceServer
	store *store.MemoryStore
}

func NewGrpcHandler(s *store.MemoryStore) *GrpcHandler {
	return &GrpcHandler{store: s}
}

func (h *GrpcHandler) CreateProfile(ctx context.Context, req *v1.CreateProfileRequest) (*v1.ProfileResponse, error) {
	if req == nil {
		return nil, apperrors.ToGRPCStatus(apperrors.ErrInvalidPayload)
	}

	payload := model.CreateProfilePayload{
		FullName: req.GetFullName(),
		Email:    req.GetEmail(),
	}

	if err := payload.Validate(); err != nil {
		return nil, apperrors.ToGRPCStatus(err)
	}

	profile, err := h.store.CreateProfile(payload.FullName, payload.Email)
	if err != nil {
		return nil, apperrors.ToGRPCStatus(err)
	}

	return &v1.ProfileResponse{Profile: profile}, nil
}

func (h *GrpcHandler) GetProfile(ctx context.Context, req *v1.GetProfileRequest) (*v1.ProfileResponse, error) {
	if req == nil || strings.TrimSpace(req.GetUserId()) == "" {
		return nil, apperrors.ToGRPCStatus(apperrors.ErrEmptyUserID)
	}

	profile, err := h.store.GetProfile(strings.TrimSpace(req.GetUserId()))
	if err != nil {
		return nil, apperrors.ToGRPCStatus(err)
	}

	return &v1.ProfileResponse{Profile: profile}, nil
}

func (h *GrpcHandler) PatchProfile(ctx context.Context, req *v1.PatchProfileRequest) (*v1.ProfileResponse, error) {
	if req == nil || strings.TrimSpace(req.GetUserId()) == "" {
		return nil, apperrors.ToGRPCStatus(apperrors.ErrEmptyUserID)
	}

	userID := strings.TrimSpace(req.GetUserId())
	var namePtr *string
	var emailPtr *string

	mask := req.GetUpdateMask()
	if mask != nil && len(mask.GetPaths()) > 0 {
		for _, path := range mask.GetPaths() {
			switch strings.TrimSpace(strings.ToLower(path)) {
			case "full_name", "fullname":
				val := req.GetFullName()
				namePtr = &val
			case "email":
				val := req.GetEmail()
				emailPtr = &val
			}
		}
	} else {
		if req.GetFullName() != "" {
			val := req.GetFullName()
			namePtr = &val
		}
		if req.GetEmail() != "" {
			val := req.GetEmail()
			emailPtr = &val
		}
	}

	patchPayload := model.PatchProfilePayload{
		FullName: namePtr,
		Email:    emailPtr,
	}

	if err := patchPayload.Validate(); err != nil {
		return nil, apperrors.ToGRPCStatus(err)
	}

	updated, err := h.store.PatchProfile(userID, patchPayload.FullName, patchPayload.Email)
	if err != nil {
		return nil, apperrors.ToGRPCStatus(err)
	}

	return &v1.ProfileResponse{Profile: updated}, nil
}

func (h *GrpcHandler) LogActivity(ctx context.Context, req *v1.LogActivityRequest) (*v1.ActivityResponse, error) {
	if req == nil {
		return nil, apperrors.ToGRPCStatus(apperrors.ErrInvalidPayload)
	}

	payload := model.LogActivityPayload{
		UserID:      req.GetUserId(),
		ActionType:  req.GetActionType(),
		Description: req.GetDescription(),
	}

	if err := payload.Validate(); err != nil {
		return nil, apperrors.ToGRPCStatus(err)
	}

	record, err := h.store.AppendActivity(payload.UserID, payload.ActionType, payload.Description)
	if err != nil {
		return nil, apperrors.ToGRPCStatus(err)
	}

	return &v1.ActivityResponse{Record: record}, nil
}

func (h *GrpcHandler) ListActivities(ctx context.Context, req *v1.ListActivitiesRequest) (*v1.ListActivitiesResponse, error) {
	if req == nil || strings.TrimSpace(req.GetUserId()) == "" {
		return nil, apperrors.ToGRPCStatus(apperrors.ErrEmptyUserID)
	}

	userID := strings.TrimSpace(req.GetUserId())
	records, err := h.store.ListActivities(userID)
	if err != nil {
		return nil, apperrors.ToGRPCStatus(err)
	}

	return &v1.ListActivitiesResponse{Activities: records}, nil
}
