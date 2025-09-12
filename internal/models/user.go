package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                int
	UUID              uuid.UUID
	Name              string
	Email             string
	EncryptedPassword string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}
