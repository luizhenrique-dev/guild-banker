package transaction

import (
	"errors"
	"net/http"
	"strconv"
	"time"

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
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req createTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	transaction, err := h.service.Create(c.Request.Context(), CreateInput{
		Type:        req.Type,
		Description: req.Description,
		Amount:      req.Amount,
		Category:    req.Category,
		Visibility:  req.Visibility,
		OccurredAt:  req.OccurredAt,
		GuildID:     guildID,
	}, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(transaction))
}

func (h *Handler) List(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	input, err := parseListInput(c, guildID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactions, nextCursor, err := h.service.List(c.Request.Context(), input, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]transactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, toResponse(transaction))
	}

	c.JSON(http.StatusOK, listTransactionsResponse{Items: items, NextCursor: nextCursor})
}

func (h *Handler) Update(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	transactionID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req updateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	transaction, err := h.service.Update(c.Request.Context(), guildID, transactionID, UpdateInput(req), actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(transaction))
}

func (h *Handler) Delete(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	transactionID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), guildID, transactionID, actor); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) BulkCategorize(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req bulkCategorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	result, err := h.service.BulkCategorize(c.Request.Context(), guildID, req.TransactionIDs, req.Category, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) SetVisibility(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	transactionID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req setVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	transaction, err := h.service.SetVisibility(c.Request.Context(), guildID, transactionID, req.Visibility, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, setVisibilityResponse{ID: transaction.ID, Visibility: transaction.Visibility, UpdatedAt: transaction.UpdatedAt})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrRequesterIsNotMember):
		c.JSON(http.StatusForbidden, gin.H{"error": "requester is not member of guild"})
	case errors.Is(err, ErrTransactionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
	case errors.Is(err, ErrTransactionCancelled):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "transaction is cancelled"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func parseListInput(c *gin.Context, guildID int64) (ListInput, error) {
	input := ListInput{GuildID: guildID, Cursor: c.Query("cursor")}

	if limitValue := c.Query("limit"); limitValue != "" {
		limit, err := strconv.Atoi(limitValue)
		if err != nil || limit <= 0 {
			return ListInput{}, errors.New("limit must be a positive integer")
		}
		input.Limit = limit
	}

	if dateFromValue := c.Query("dateFrom"); dateFromValue != "" {
		dateFrom, err := time.Parse(time.RFC3339, dateFromValue)
		if err != nil {
			return ListInput{}, errors.New("dateFrom must be RFC3339")
		}
		input.DateFrom = &dateFrom
	}
	if dateToValue := c.Query("dateTo"); dateToValue != "" {
		dateTo, err := time.Parse(time.RFC3339, dateToValue)
		if err != nil {
			return ListInput{}, errors.New("dateTo must be RFC3339")
		}
		input.DateTo = &dateTo
	}

	if categoryValue := c.Query("category"); categoryValue != "" {
		category := Category(categoryValue)
		input.Category = &category
	}
	if typeValue := c.Query("type"); typeValue != "" {
		typeEnum := Type(typeValue)
		input.Type = &typeEnum
	}
	if sourceValue := c.Query("source"); sourceValue != "" {
		source := Source(sourceValue)
		input.Source = &source
	}
	if statusValue := c.Query("status"); statusValue != "" {
		status := Status(statusValue)
		input.Status = &status
	}
	if visibilityValue := c.Query("visibility"); visibilityValue != "" {
		visibility := Visibility(visibilityValue)
		input.Visibility = &visibility
	}

	return input, nil
}
