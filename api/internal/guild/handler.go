package guild

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createGuildRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type updateGuildRequest struct {
	Name string `json:"name"`
}

type inviteUserRequest struct {
	Email string `json:"email"`
}

type guildResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
}

func (h *Handler) Create(c *gin.Context) {
	var req createGuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	g, err := h.service.Create(c.Request.Context(), req.Name, req.DisplayName, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, guildResponse{ID: g.ID, Name: g.Name, DisplayName: g.DisplayName, Enabled: g.Enabled})
}

func (h *Handler) UpdateName(c *gin.Context) {
	guildID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req updateGuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	g, err := h.service.UpdateName(c.Request.Context(), guildID, req.Name, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, guildResponse{ID: g.ID, Name: g.Name, DisplayName: g.DisplayName, Enabled: g.Enabled})
}

func (h *Handler) Enable(c *gin.Context) {
	h.updateEnabledStatus(c, true)
}

func (h *Handler) Disable(c *gin.Context) {
	h.updateEnabledStatus(c, false)
}

func (h *Handler) ListByMember(c *gin.Context) {
	requesterUserID, ok := getRequesterUserID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id header is required"})
		return
	}

	guilds, err := h.service.ListByMember(c.Request.Context(), requesterUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := make([]guildResponse, 0, len(guilds))
	for _, g := range guilds {
		resp = append(resp, guildResponse{ID: g.ID, Name: g.Name, DisplayName: g.DisplayName, Enabled: g.Enabled})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) InviteUser(c *gin.Context) {
	guildID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req inviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	requesterUserID, ok := getRequesterUserID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id header is required"})
		return
	}

	if err := h.service.InviteUser(c.Request.Context(), requesterUserID, guildID, req.Email); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) RemoveUser(c *gin.Context) {
	guildID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	memberID, err := parseIDParam(c, "userID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requesterUserID, ok := getRequesterUserID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id header is required"})
		return
	}

	if err := h.service.RemoveUser(c.Request.Context(), requesterUserID, guildID, memberID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) updateEnabledStatus(c *gin.Context, enable bool) {
	guildID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	if enable {
		err = h.service.Enable(c.Request.Context(), guildID, actor)
	} else {
		err = h.service.Disable(c.Request.Context(), guildID, actor)
	}
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrGuildNameAlreadyUsed):
		c.JSON(http.StatusConflict, gin.H{"error": "guild name already used"})
	case errors.Is(err, ErrRequesterIsNotMember):
		c.JSON(http.StatusForbidden, gin.H{"error": "requester is not a member of this guild"})
	case errors.Is(err, ErrGuildNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "guild not found"})
	case errors.Is(err, ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	case errors.Is(err, ErrGuildMemberNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "guild member not found"})
	case errors.Is(err, ErrGuildMemberAlreadySet):
		c.JSON(http.StatusConflict, gin.H{"error": "user is already a guild member"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func parseIDParam(c *gin.Context, name string) (int64, error) {
	value := c.Param(name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New(name + " is invalid")
	}
	return id, nil
}

func getRequesterActor(c *gin.Context) (audit.Actor, bool) {
	userID, ok := getRequesterUserID(c)
	if !ok {
		return audit.Actor{}, false
	}
	email := c.GetHeader("X-User-Email")
	if email == "" {
		return audit.Actor{}, false
	}
	return audit.Actor{UserID: userID, Email: email}, true
}

func getRequesterUserID(c *gin.Context) (int64, bool) {
	value := c.GetHeader("X-User-ID")
	if value == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}
