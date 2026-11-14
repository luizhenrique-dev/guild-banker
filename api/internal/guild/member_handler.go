package guild

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type memberResponse struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	LastAccessAt *time.Time `json:"last_access_at"`
	MemberSince  time.Time  `json:"member_since"`
}

type sentInvitationResponse struct {
	ID        int64           `json:"id"`
	Email     string          `json:"email"`
	InvitedBy inviterResponse `json:"invited_by"`
	CreatedAt time.Time       `json:"created_at"`
}

func (h *Handler) ListMembers(c *gin.Context) {
	requesterID, guildID, ok := memberRequest(c)
	if !ok {
		return
	}
	members, err := h.service.ListMembers(c.Request.Context(), requesterID, guildID)
	if err != nil {
		h.handleInvitationError(c, err)
		return
	}
	response := make([]memberResponse, 0, len(members))
	for _, m := range members {
		response = append(response, memberResponse(m))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) ListSentInvitations(c *gin.Context) {
	requesterID, guildID, ok := memberRequest(c)
	if !ok {
		return
	}
	invitations, err := h.service.ListSentInvitations(c.Request.Context(), requesterID, guildID)
	if err != nil {
		h.handleInvitationError(c, err)
		return
	}
	response := make([]sentInvitationResponse, 0, len(invitations))
	for _, i := range invitations {
		response = append(response, sentInvitationResponse{ID: i.ID, Email: i.Email, CreatedAt: i.CreatedAt, InvitedBy: inviterResponse{Name: i.InvitedByName, Email: i.InvitedByEmail}})
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) CancelInvitation(c *gin.Context) {
	requesterID, guildID, ok := memberRequest(c)
	if !ok {
		return
	}
	id, err := parseIDParam(c, "inviteID")
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation id"})
		return
	}
	if err := h.service.CancelInvitation(c.Request.Context(), requesterID, guildID, id); err != nil {
		h.handleInvitationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func memberRequest(c *gin.Context) (int64, int64, bool) {
	requesterID, ok := getRequesterUserID(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return 0, 0, false
	}
	guildID, err := parseIDParam(c, "guildID")
	if err != nil || guildID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guild id"})
		return 0, 0, false
	}
	return requesterID, guildID, true
}
