package handlers

import (
	"slot-booking-service-api/internal/usecase/auth"
	"slot-booking-service-api/internal/usecase/user"
)

type UserHandler struct {
	UserUC *user.UserUseCase
	AuthUC *auth.AuthUseCase
}

func NewUserHandler(UserUseCase *user.UserUseCase, AuthUseCase *auth.AuthUseCase) *UserHandler {
	return &UserHandler{
		UserUC: UserUseCase,
		AuthUC: AuthUseCase,
	}
}
