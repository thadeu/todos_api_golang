package response

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	UUID      string    `json:"uuid,omitempty"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type TodoResponse struct {
	UUID        uuid.UUID `json:"uuid"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status,omitempty"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CursorData struct {
	Datetime string `json:"datetime"`
	ID       int    `json:"id,omitempty"`
}

type CursorResponse struct {
	Size       int             `json:"size"`
	Data       json.RawMessage `json:"data"`
	Pagination struct {
		HasNext    bool   `json:"has_next"`
		NextCursor string `json:"next_cursor"`
	} `json:"pagination"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ResponseError struct {
	Code    string            `json:"code"`
	Errors  []ValidationError `json:"errors"`
	Details any               `json:"details,omitempty"`
}

type SuccessResponse struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error ResponseError `json:"error"`
}
