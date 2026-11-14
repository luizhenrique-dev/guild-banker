package fixedexpense

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var req createFixedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fixedExpense, err := h.service.CreateWithResponsible(c.Request.Context(), guildID, req.Name, req.Amount, req.DueDay, req.Category, req.ResponsibleUserID, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(fixedExpense))
}

func (h *Handler) ListActiveByGuild(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var includePaused bool
	if value, present := c.GetQuery("include_paused"); present {
		includePaused, err = strconv.ParseBool(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid include_paused parameter"})
			return
		}
	}
	var fixedExpenses []*FixedExpense
	if includePaused {
		fixedExpenses, err = h.service.ListByGuild(c.Request.Context(), guildID, actor)
	} else {
		fixedExpenses, err = h.service.ListActiveByGuild(c.Request.Context(), guildID, actor)
	}
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
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fixedExpense, err := h.service.Update(c.Request.Context(), guildID, id, UpdateInput{
		Amount: req.Amount, DueDay: req.DueDay, Category: req.Category, Status: req.Status,
		ResponsibleUserID: req.ResponsibleUserID.value, SetResponsible: req.ResponsibleUserID.present,
	}, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(fixedExpense))
}

func (h *Handler) Deactivate(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.service.Deactivate(c.Request.Context(), guildID, id, req.Status, actor); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), guildID, id, actor); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrRequesterIsNotMember):
		c.JSON(http.StatusForbidden, gin.H{"error": "requester is not a member of this guild"})
	case errors.Is(err, ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fixed expense input"})
	case errors.Is(err, ErrFixedExpenseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "fixed expense not found"})
	case errors.Is(err, ErrInvalidFixedExpenseStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fixed expense status"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	principal, ok := auth.PrincipalFromContext(c.Request.Context())
	return principal.Actor, ok
}
