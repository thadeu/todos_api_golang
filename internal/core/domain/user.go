package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	Admin   UserRole = "admin"
	Profile UserRole = "profile"
)

type User struct {
	ID                int
	UUID              uuid.UUID
	Name              string `validate:"required,min=2,max=100"`
	Email             string `validate:"required,email,max=255"`
	EncryptedPassword string `validate:"required"`
	Role              UserRole
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}
