package user

import (
	"context"
	"errors"
	"slot-booking-service-api/internal/entities/user"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	Id string

	Name         string
	Email        string
	PasswordHash string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func ToUserRecord(ue *user.User) *User {
	return &User{
		Id:           ue.Id,
		Name:         ue.Name,
		Email:        ue.Email,
		PasswordHash: ue.PasswordHash,
		CreatedAt:    ue.CreatedAt,
		UpdatedAt:    ue.UpdatedAt,
	}
}

func ToUserEntity(ur *User) *user.User {
	return &user.User{
		Id:           ur.Id,
		Name:         ur.Name,
		Email:        ur.Email,
		PasswordHash: ur.PasswordHash,
		CreatedAt:    ur.CreatedAt,
		UpdatedAt:    ur.UpdatedAt,
	}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	_, err := r.db.Exec(ctx, CREATE,
		user.Id,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetById(ctx context.Context, id string) (*User, error) {
	var u User

	err := r.db.QueryRow(ctx, GET_BY_ID, id).Scan(
		&u.Id,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User

	err := r.db.QueryRow(ctx, GET_BY_EMAIL, email).Scan(
		&u.Id,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}
