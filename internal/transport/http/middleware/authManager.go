package middleware

import "slot-booking-service-api/internal/usecase/auth"

type AuthManager struct {
	AuthUC *auth.AuthUseCase
}

func New(authUC *auth.AuthUseCase) *AuthManager {
	return &AuthManager{
		AuthUC: authUC,
	}
}
