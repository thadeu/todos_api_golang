package request

import "time"

type SignUpRequest struct {
	Email    string `json:"email,omitempty" validate:"required,email,max=255"`
	Password string `json:"password,omitempty" validate:"required,min=6,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email,omitempty" validate:"required,email,max=255"`
	Password string `json:"password,omitempty" validate:"required,min=6,max=100"`
}

type TodoRequest struct {
	Title       string     `json:"title,omitempty" validate:"min=3,max=255"`
	Description string     `json:"description,omitempty" validate:"max=1000"`
	Status      string     `json:"status,omitempty"`
	Completed   bool       `json:"completed,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type UserRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}
