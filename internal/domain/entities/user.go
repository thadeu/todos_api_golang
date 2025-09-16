package entities

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity in the domain
type User struct {
	ID                int
	UUID              uuid.UUID
	Name              string `validate:"required,min=2,max=100"`
	Email             string `validate:"required,email,max=255"`
	EncryptedPassword string `validate:"required"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// IsDeleted checks if the user is marked as deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}
