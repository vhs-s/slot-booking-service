package dto

import (
	"slot-booking-service-api/internal/entities/user"
	"time"
)

type GetUserDto struct {
	Id string

	Name  string
	Email string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func ToGetUserDto(ue *user.User) *GetUserDto {
	return &GetUserDto{
		Id:        ue.Id, // Если в entity ID типа int/uuid, добавь конвертацию (например, strconv.Itoa или .String())
		Name:      ue.Name,
		Email:     ue.Email,
		CreatedAt: ue.CreatedAt,
		UpdatedAt: ue.UpdatedAt,
	}
}
