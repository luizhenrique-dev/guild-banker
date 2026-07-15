package importer

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

const maxUploadSizeBytes int64 = 5 * 1024 * 1024 // 5MB

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Upload(c *gin.Context) {
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

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if file.Size <= 0 || file.Size > maxUploadSizeBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file size"})
		return
	}

	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open file"})
		return
	}
	defer func() { _ = opened.Close() }()

	batch, items, summary, err := h.service.Upload(c.Request.Context(), UploadInput{
		GuildID:    guildID,
		FileName:   file.Filename,
		File:       opened,
		SourceBank: SourceBankC6,
	}, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	responseItems := make([]itemResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, toItemResponse(item))
	}

	c.JSON(http.StatusCreated, uploadResponse{
		ImportID: batch.ID,
		Status:   batch.Status,
		FileName: batch.FileName,
		Summary:  summary,
		Items:    responseItems,
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	importID, err := parseIDParam(c, "importID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	batch, items, err := h.service.GetByID(c.Request.Context(), guildID, importID, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	responseItems := make([]itemResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, toItemResponse(item))
	}

	c.JSON(http.StatusOK, getBatchResponse{ImportID: batch.ID, Status: batch.Status, FileName: batch.FileName, Items: responseItems})
}

func (h *Handler) UpdateItem(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	importID, err := parseIDParam(c, "importID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	itemID, err := parseIDParam(c, "itemID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req updateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	item, err := h.service.UpdateItem(c.Request.Context(), UpdateItemInput{
		GuildID:     guildID,
		ImportID:    importID,
		ItemID:      itemID,
		Description: req.Description,
		OccurredAt:  req.OccurredAt,
		Amount:      req.Amount,
		Type:        req.Type,
		Category:    req.Category,
	}, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toItemResponse(item))
}

func (h *Handler) DeleteItem(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	importID, err := parseIDParam(c, "importID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	itemID, err := parseIDParam(c, "itemID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	if err := h.service.DiscardItem(c.Request.Context(), guildID, importID, itemID, actor); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Confirm(c *gin.Context) {
	guildID, err := parseIDParam(c, "guildID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	importID, err := parseIDParam(c, "importID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actor, ok := getRequesterActor(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x-user-id and x-user-email headers are required"})
		return
	}

	created, skipped, err := h.service.Confirm(c.Request.Context(), ConfirmInput{GuildID: guildID, ImportID: importID}, actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, confirmResponse{ImportID: importID, Created: created, Skipped: skipped, Status: BatchStatusCompleted})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrRequesterIsNotMember):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrImportBatchNotFound), errors.Is(err, ErrImportItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func parseIDParam(c *gin.Context, name string) (int64, error) {
	value := c.Param(name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid " + name)
	}

	return id, nil
}

func getRequesterActor(c *gin.Context) (audit.Actor, bool) {
	userID, ok := getRequesterUserID(c)
	if !ok {
		return audit.Actor{}, false
	}

	email := c.GetHeader("x-user-email")
	if email == "" {
		return audit.Actor{}, false
	}

	return audit.Actor{UserID: userID, Email: email}, true
}

func getRequesterUserID(c *gin.Context) (int64, bool) {
	value := c.GetHeader("x-user-id")
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}
