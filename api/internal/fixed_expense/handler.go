package fixedexpense

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

func (h *Handler) Create(c *gin.Context) {
	var req createFixedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	fixedExpense, err := h.service.Create(c.Request.Context(), req.Name, req.Amount, req.DueDay, req.Category, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(fixedExpense))
}

func (h *Handler) ListActiveByUser(c *gin.Context) {
	requesterUserID, ok := getRequesterUserID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id header is required"})
		return
	}

	fixedExpenses, err := h.service.ListActiveByUser(c.Request.Context(), requesterUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := make([]fixedExpenseResponse, 0, len(fixedExpenses))
	for _, fixedExpense := range fixedExpenses {
		resp = append(resp, toResponse(fixedExpense))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req updateFixedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	fixedExpense, err := h.service.Update(c.Request.Context(), id, UpdateInput(req), actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(fixedExpense))
}

func (h *Handler) Deactivate(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req deactivateFixedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	if err := h.service.Deactivate(c.Request.Context(), id, req.Status, actor); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrFixedExpenseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "fixed expense not found"})
	case errors.Is(err, ErrInvalidFixedExpenseStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fixed expense status"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func parseIDParam(c *gin.Context, name string) (int64, error) {
	value := c.Param(name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New(name + " must be a positive integer")
	}

	return id, nil
}

func getRequesterActor(c *gin.Context) (audit.Actor, bool) {
	requesterUserID, ok := getRequesterUserID(c)
	if !ok {
		return audit.Actor{}, false
	}

	email := c.GetHeader("x-user-email")
	if email == "" {
		return audit.Actor{}, false
	}

	return audit.Actor{UserID: requesterUserID, Email: email}, true
}

func getRequesterUserID(c *gin.Context) (int64, bool) {
	value := c.GetHeader("x-user-id")
	userID, err := strconv.ParseInt(value, 10, 64)
	if err != nil || userID <= 0 {
		return 0, false
	}

	return userID, true
}
