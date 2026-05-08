package handlers

// UpdateProfileRequest represents profile update payload
type UpdateProfileRequest struct {
	FirstName       *string `json:"first_name,omitempty"`
	LastName        *string `json:"last_name,omitempty"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	Language        *string `json:"language,omitempty"`
	IsPublic        *bool   `json:"is_public,omitempty"`
	FavoritesPublic *bool   `json:"favorites_public,omitempty"`
}
