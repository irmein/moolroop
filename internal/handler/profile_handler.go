package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/moolroop/gen/v1"
	apperrors "github.com/moolroop/internal/errors"
	"github.com/moolroop/internal/model"
	"github.com/moolroop/internal/store"
)

type ProfileHandler struct {
	store *store.MemoryStore
}

func NewProfileHandler(s *store.MemoryStore) *ProfileHandler {
	return &ProfileHandler{store: s}
}

// CreateProfile godoc
// @Summary      Create a new user profile
// @Description  Creates a new user profile with minimal PII (full name and email only)
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        X-Request-ID  header    string                      false  "Optional client-supplied correlation request ID"
// @Param        payload       body      model.CreateProfilePayload  true   "Create Profile Request"
// @Success      201           {object}  v1.UserProfile
// @Failure      400           {object}  model.ErrorResponse
// @Failure      500           {object}  model.ErrorResponse
// @Header       201           {string}  X-Request-ID  "Unique correlation request ID"
// @Router       /profiles [post]
func (h *ProfileHandler) CreateProfile(c *gin.Context) {
	var payload model.CreateProfilePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid payload: " + err.Error()})
		return
	}

	if err := payload.Validate(); err != nil {
		c.JSON(apperrors.ToHTTPStatus(err), model.ErrorResponse{Error: err.Error()})
		return
	}

	profile, err := h.store.CreateProfile(payload.FullName, payload.Email)
	if err != nil {
		c.JSON(apperrors.ToHTTPStatus(err), model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, profile)
}

// GetProfile godoc
// @Summary      Get user profile by ID
// @Description  Retrieves profile information by user ID
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        X-Request-ID  header    string  false  "Optional client-supplied correlation request ID"
// @Param        id            path      string  true   "User ID"
// @Success      200           {object}  v1.UserProfile
// @Failure      400           {object}  model.ErrorResponse
// @Failure      404           {object}  model.ErrorResponse
// @Failure      500           {object}  model.ErrorResponse
// @Header       200           {string}  X-Request-ID  "Unique correlation request ID"
// @Router       /profiles/{id} [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "user id is required"})
		return
	}

	profile, err := h.store.GetProfile(id)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to retrieve profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// PatchProfile godoc
// @Summary      Patch an existing user profile
// @Description  Incrementally update profile fields without rewriting full entity
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        X-Request-ID  header    string                     false  "Optional client-supplied correlation request ID"
// @Param        id            path      string                     true   "User ID"
// @Param        payload       body      model.PatchProfilePayload  true   "Fields to patch"
// @Success      200           {object}  v1.UserProfile
// @Failure      400           {object}  model.ErrorResponse
// @Failure      404           {object}  model.ErrorResponse
// @Failure      500           {object}  model.ErrorResponse
// @Header       200           {string}  X-Request-ID  "Unique correlation request ID"
// @Router       /profiles/{id} [patch]
func (h *ProfileHandler) PatchProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "user id is required"})
		return
	}

	var payload model.PatchProfilePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid payload: " + err.Error()})
		return
	}

	if err := payload.Validate(); err != nil {
		c.JSON(apperrors.ToHTTPStatus(err), model.ErrorResponse{Error: err.Error()})
		return
	}

	updated, err := h.store.PatchProfile(id, payload.FullName, payload.Email)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to patch profile"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
