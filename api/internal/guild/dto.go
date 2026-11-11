package guild

type createGuildRequest struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type updateGuildRequest struct {
	DisplayName string  `json:"display_name"`
	Description *string `json:"description"`
}

type inviteUserRequest struct {
	Email string `json:"email"`
}

type guildResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}
