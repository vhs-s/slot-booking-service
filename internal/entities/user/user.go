package user

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id string

	Name         string
	Email        string
	PasswordHash string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(name, email, password string) *User {
	h := sha256.New()
	h.Write([]byte(password))
	hashedData := h.Sum(nil)
	hexHash := hex.EncodeToString(hashedData)
	return &User{
		Id:           uuid.NewString(),
		Name:         name,
		Email:        email,
		PasswordHash: hexHash,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}
