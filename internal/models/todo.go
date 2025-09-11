package models

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID          int
	UUID        uuid.UUID
	Title       string
	Description string
	Completed   bool
	UserId      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
