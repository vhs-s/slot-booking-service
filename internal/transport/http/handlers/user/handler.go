package handlers

import (
	"slot-booking-service-api/internal/usecase/auth"
	"slot-booking-service-api/internal/usecase/booking"
	"slot-booking-service-api/internal/usecase/schedule"
	"slot-booking-service-api/internal/usecase/user"
)

type UserHandler struct {
	UserUC     *user.UserUseCase
	AuthUC     *auth.AuthUseCase
	ScheduleUC *schedule.ScheduleUseCase
	BookingUC  *booking.BookingUseCase
}

func NewUserHandler(UserUseCase *user.UserUseCase, AuthUseCase *auth.AuthUseCase, scheduleUseCases ...*schedule.ScheduleUseCase) *UserHandler {
	handler := &UserHandler{
		UserUC: UserUseCase,
		AuthUC: AuthUseCase,
	}
	if len(scheduleUseCases) > 0 {
		handler.ScheduleUC = scheduleUseCases[0]
	}
	return handler
}
