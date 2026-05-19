package types

// RegisterRequest represents user registration payload
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=30"`
	Email     string `json:"email" binding:"required,email,max=254"`
	Password  string `json:"password" binding:"required,min=8,max=128"`
	FirstName string `json:"first_name" binding:"required,min=1,max=50"`
	LastName  string `json:"last_name" binding:"required,min=1,max=50"`
}

// LoginRequest represents user login payload
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AvailabilityRequest represents availability check payload
type AvailabilityRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}

// AvailabilityResponse represents the response for availability checks
type AvailabilityResponse struct {
	Status      string   `json:"status"`
	Available   bool     `json:"available"`
	Message     string   `json:"message,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// EmailVerificationRequest represents email verification request
type EmailVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyEmailRequest represents email verification code request
type VerifyEmailRequest struct {
	Email            string `json:"email" binding:"required,email"`
	VerificationCode string `json:"verification_code" binding:"required"`
}
