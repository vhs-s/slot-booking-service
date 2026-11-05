package handlers

import (
	bookingusecase "slot-booking-service-api/internal/usecase/booking"
	userusecase "slot-booking-service-api/internal/usecase/user"
)

type BookingHandler struct {
	UserUC    *userusecase.UserUseCase
	BookingUC *bookingusecase.BookingUseCase
}

func NewBookingHandler(userUC *userusecase.UserUseCase, bookingUC *bookingusecase.BookingUseCase) *BookingHandler {
	return &BookingHandler{UserUC: userUC, BookingUC: bookingUC}
}
