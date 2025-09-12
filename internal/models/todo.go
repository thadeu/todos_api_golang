package models

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID          int
	UUID        uuid.UUID
	Title       string `validate:"required,min=4,max=255"`
	Description string `validate:"max=255"`
	Status      int
	Completed   bool `validate:"boolean"`
	UserId      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
