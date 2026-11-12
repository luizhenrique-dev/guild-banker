package guild

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type inviterResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type invitationResponse struct {
	ID        int64           `json:"id"`
	Guild     guildResponse   `json:"guild"`
	InvitedBy inviterResponse `json:"invited_by"`
}

func (h *Handler) ListInvitations(c *gin.Context) {
	actor, ok := getRequesterActor(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}
	invitations, err := h.service.ListInvitations(c.Request.Context(), actor)
	if err != nil {
		h.handleInvitationError(c, err)
		return
	}
	response := make([]invitationResponse, 0, len(invitations))
	for _, invitation := range invitations {
		g := invitation.Guild
		response = append(response, invitationResponse{ID: invitation.ID, Guild: guildResponse{ID: g.ID, Name: g.Name, DisplayName: g.DisplayName, Description: g.Description, Enabled: g.Enabled}, InvitedBy: inviterResponse{Name: invitation.InvitedByName, Email: invitation.InvitedByEmail}})
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) AcceptInvitation(c *gin.Context)  { h.respondInvitation(c, true) }
func (h *Handler) DeclineInvitation(c *gin.Context) { h.respondInvitation(c, false) }
func (h *Handler) respondInvitation(c *gin.Context, accept bool) {
	actor, ok := getRequesterActor(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}
	id, err := parseIDParam(c, "inviteID")
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation id"})
		return
	}
	g, err := h.service.RespondInvitation(c.Request.Context(), id, actor, accept)
	if err != nil {
		h.handleInvitationError(c, err)
		return
	}
	if !accept {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, guildResponse{ID: g.ID, Name: g.Name, DisplayName: g.DisplayName, Description: g.Description, Enabled: g.Enabled})
}
