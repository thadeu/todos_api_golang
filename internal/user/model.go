package user

import (
	"time"

	"github.com/google/uuid"
)

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
