package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "moolroop/gen/v1"
	apperrors "moolroop/internal/errors"
	"moolroop/internal/model"
	"moolroop/internal/store"
)

type ActivityHandler struct {
	store *store.MemoryStore
}

func NewActivityHandler(s *store.MemoryStore) *ActivityHandler {
	return &ActivityHandler{store: s}
}

// LogActivity godoc
// @Summary      Log an immutable activity record
// @Description  Creates an append-only, immutable activity log event for a user
// @Tags         activity
// @Accept       json
// @Produce      json
// @Param        payload  body      model.LogActivityPayload  true  "Activity payload"
// @Success      201      {object}  v1.ActivityRecord
// @Failure      400      {object}  gin.H
// @Failure      404      {object}  gin.H
// @Failure      500      {object}  gin.H
// @Router       /activities [post]
func (h *ActivityHandler) LogActivity(c *gin.Context) {
	var payload model.LogActivityPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	if err := payload.Validate(); err != nil {
		c.JSON(apperrors.ToHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	record, err := h.store.AppendActivity(payload.UserID, payload.ActionType, payload.Description)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log activity"})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// ListActivities godoc
// @Summary      List immutable activity history for a user
// @Description  Retrieves all historical activity records for the specified user
// @Tags         activity
// @Accept       json
// @Produce      json
// @Param        user_id  path      string  true  "User ID"
// @Success      200      {array}   v1.ActivityRecord
// @Failure      400      {object}  gin.H
// @Failure      404      {object}  gin.H
// @Failure      500      {object}  gin.H
// @Router       /activities/{user_id} [get]
func (h *ActivityHandler) ListActivities(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	records, err := h.store.ListActivities(userID)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list activities"})
		return
	}

	c.JSON(http.StatusOK, records)
}
