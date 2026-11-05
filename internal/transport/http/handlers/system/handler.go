package handlers

import "slot-booking-service-api/internal/usecase/user"

type SystemHandler struct {
	UserUC *user.UserUseCase
}

func NewSystemHandler(userUC *user.UserUseCase) *SystemHandler {
	return &SystemHandler{UserUC: userUC}
}
