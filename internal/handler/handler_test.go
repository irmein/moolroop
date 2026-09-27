package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/moolroop/gen/v1"
	"github.com/moolroop/internal/handler"
	"github.com/moolroop/internal/middleware"
	"github.com/moolroop/internal/model"
	"github.com/moolroop/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func setupTestRouter(s *store.MemoryStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())

	profileHandler := handler.NewProfileHandler(s)
	activityHandler := handler.NewActivityHandler(s)

	v1Group := r.Group("/api/v1")
	{
		v1Group.POST("/profiles", profileHandler.CreateProfile)
		v1Group.GET("/profiles/:id", profileHandler.GetProfile)
		v1Group.PATCH("/profiles/:id", profileHandler.PatchProfile)

		v1Group.POST("/activities", activityHandler.LogActivity)
		v1Group.GET("/activities/:user_id", activityHandler.ListActivities)
	}

	return r
}

func TestProfileRoutes(t *testing.T) {
	s := store.NewMemoryStore()
	r := setupTestRouter(s)

	// 1. Create Profile
	createPayload := model.CreateProfilePayload{
		FullName: "Charlie Brown",
		Email:    "charlie@example.com",
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	var createdProfile v1.UserProfile
	err := json.Unmarshal(w.Body.Bytes(), &createdProfile)
	require.NoError(t, err)
	assert.NotEmpty(t, createdProfile.UserId)
	assert.Equal(t, "Charlie Brown", createdProfile.FullName)
	assert.Equal(t, "charlie@example.com", createdProfile.Email)

	// 2. Create Profile - Invalid Email
	badEmailPayload := model.CreateProfilePayload{
		FullName: "Bad Email",
		Email:    "invalid-email",
	}
	body, _ = json.Marshal(badEmailPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var errResp model.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(errResp.Error), "email")

	// 3. Get Profile
	req = httptest.NewRequest(http.MethodGet, "/api/v1/profiles/"+createdProfile.UserId, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))

	// 4. Get Profile - Not Found
	req = httptest.NewRequest(http.MethodGet, "/api/v1/profiles/non-existing-id", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "user profile not found", errResp.Error)

	// 5. Patch Profile
	newName := "Charles Brown"
	patchPayload := model.PatchProfilePayload{
		FullName: &newName,
	}
	body, _ = json.Marshal(patchPayload)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/profiles/"+createdProfile.UserId, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	var patchedProfile v1.UserProfile
	err = json.Unmarshal(w.Body.Bytes(), &patchedProfile)
	require.NoError(t, err)
	assert.Equal(t, "Charles Brown", patchedProfile.FullName)
	assert.Equal(t, "charlie@example.com", patchedProfile.Email)

	// 6. Patch Profile - Not Found
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/profiles/unknown-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestActivityRoutes(t *testing.T) {
	s := store.NewMemoryStore()
	r := setupTestRouter(s)

	// Create user first
	p, err := s.CreateProfile("David", "david@example.com")
	require.NoError(t, err)

	// 1. Log Activity
	actPayload := model.LogActivityPayload{
		UserID:      p.UserId,
		ActionType:  "LOGIN_SUCCESS",
		Description: "User logged in via WebAuthn",
	}
	body, _ := json.Marshal(actPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/activities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	var record v1.ActivityRecord
	err = json.Unmarshal(w.Body.Bytes(), &record)
	require.NoError(t, err)
	assert.NotEmpty(t, record.ActivityId)
	assert.Equal(t, "LOGIN_SUCCESS", record.ActionType)

	// 2. Log Activity - User Not Found
	actBadUser := model.LogActivityPayload{
		UserID:      "unknown-user-id",
		ActionType:  "TEST",
		Description: "test",
	}
	body, _ = json.Marshal(actBadUser)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/activities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 3. List Activities
	req = httptest.NewRequest(http.MethodGet, "/api/v1/activities/"+p.UserId, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	var list []v1.ActivityRecord
	err = json.Unmarshal(w.Body.Bytes(), &list)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// 4. List Activities - User Not Found
	req = httptest.NewRequest(http.MethodGet, "/api/v1/activities/unknown-id", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGrpcHandler(t *testing.T) {
	s := store.NewMemoryStore()
	h := handler.NewGrpcHandler(s)
	ctx := context.Background()

	// 1. Create Profile
	createResp, err := h.CreateProfile(ctx, &v1.CreateProfileRequest{
		FullName: "Eva Green",
		Email:    "eva@example.com",
	})
	require.NoError(t, err)
	require.NotNil(t, createResp.Profile)
	userID := createResp.Profile.UserId
	assert.Equal(t, "Eva Green", createResp.Profile.FullName)

	// 2. Create Profile - Invalid Email
	_, err = h.CreateProfile(ctx, &v1.CreateProfileRequest{
		FullName: "Eva Green",
		Email:    "not-an-email",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())

	// 3. Get Profile
	getResp, err := h.GetProfile(ctx, &v1.GetProfileRequest{UserId: userID})
	require.NoError(t, err)
	assert.Equal(t, userID, getResp.Profile.UserId)

	// 4. Get Profile - Not Found
	_, err = h.GetProfile(ctx, &v1.GetProfileRequest{UserId: "non-existent"})
	require.Error(t, err)
	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())

	// 5. Patch Profile
	mask, err := fieldmaskpb.New(&v1.UserProfile{}, "full_name")
	require.NoError(t, err)
	patchResp, err := h.PatchProfile(ctx, &v1.PatchProfileRequest{
		UserId:     userID,
		FullName:   "Eva G.",
		UpdateMask: mask,
	})
	require.NoError(t, err)
	assert.Equal(t, "Eva G.", patchResp.Profile.FullName)
	assert.Equal(t, "eva@example.com", patchResp.Profile.Email)

	// 6. Log Activity
	logResp, err := h.LogActivity(ctx, &v1.LogActivityRequest{
		UserId:      userID,
		ActionType:  "PASSWORD_RESET",
		Description: "Requested reset link",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, logResp.Record.ActivityId)

	// 7. List Activities
	listResp, err := h.ListActivities(ctx, &v1.ListActivitiesRequest{UserId: userID})
	require.NoError(t, err)
	assert.Len(t, listResp.Activities, 1)
}
