package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	c "todoapp/pkg/db/cursor"
	. "todoapp/pkg/tracing"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"todoapp/internal/domain/entities"
	"todoapp/internal/domain/repositories"
)

// todoRepository implements the TodoRepository interface
type todoRepository struct {
	db *sql.DB
}

// NewTodoRepository creates a new todo repository
func NewTodoRepository(db *sql.DB) repositories.TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(ctx context.Context, todo entities.Todo) (entities.Todo, error) {
	stmt, err := r.db.PrepareContext(ctx, "INSERT INTO todos (uuid, title, description, status, completed, user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		return entities.Todo{}, err
	}

	defer stmt.Close()

	uuid := todo.UUID.String()
	status := 0

	if todo.Status != 0 {
		status = todo.Status
	}

	_, err = stmt.ExecContext(ctx,
		uuid,
		todo.Title,
		todo.Description,
		status,
		todo.Completed,
		todo.UserId,
		todo.CreatedAt,
		todo.UpdatedAt,
	)

	if err != nil {
		slog.Error("Error creating todo", "error", err)
		return entities.Todo{}, err
	}

	saved, err := r.GetByUUID(ctx, uuid, todo.UserId)

	if err != nil {
		return entities.Todo{}, err
	}

	return saved, nil
}

func (r *todoRepository) GetAllWithCursor(ctx context.Context, userId int, limit int, cursor string) ([]entities.Todo, bool, error) {
	// Create span for database operation
	ctx, span := CreateChildSpan(ctx, "db.todo.GetAllWithCursor", []attribute.KeyValue{
		attribute.String("db.table", "todos"),
		attribute.String("db.operation", "SELECT"),
		attribute.Int("user.id", userId),
		attribute.Int("todo.limit", limit),
		attribute.String("todo.cursor", cursor),
	})
	defer span.End()

	actualLimit := limit + 1

	var query string
	var args []interface{}

	if cursor == "" {
		query = "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at FROM todos WHERE user_id = ? AND deleted_at IS NULL ORDER BY created_at DESC, id DESC LIMIT ?"
		args = []interface{}{userId, actualLimit}
	} else {
		datetimeStr, id, err := c.DecodeCursor(cursor)

		if err != nil {
			AddSpanError(span, err)
			slog.Error("Error decoding cursor", "error", err)
			return []entities.Todo{}, false, err
		}

		// Parse the datetime string to ensure proper comparison
		datetime, err := time.Parse(time.RFC3339, datetimeStr)

		if err != nil {
			slog.Error("Error parsing cursor datetime", "error", err, "datetime", datetimeStr)
			return []entities.Todo{}, false, err
		}

		query = "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at FROM todos WHERE user_id = ? AND (created_at < ? OR (created_at = ? AND id < ?)) AND deleted_at IS NULL ORDER BY created_at DESC, id DESC LIMIT ?"
		args = []interface{}{userId, datetime, datetime, id, actualLimit}
	}

	stmt, err := r.db.PrepareContext(ctx, query)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []entities.Todo{}, false, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []entities.Todo{}, false, err
	}

	defer rows.Close()

	data := []entities.Todo{}

	for rows.Next() {
		var todo entities.Todo

		err = rows.Scan(&todo.ID, &todo.UUID, &todo.Title, &todo.Description, &todo.Status, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt)

		if err != nil {
			return []entities.Todo{}, false, err
		}

		data = append(data, todo)
	}

	// Check if there's a next page
	hasNext := len(data) == actualLimit

	// If there are more data than limit, remove the extra item
	if hasNext {
		data = data[:limit]
	}

	// Add success attributes
	span.SetAttributes(
		attribute.Int("db.rows_returned", len(data)),
		attribute.Bool("db.has_next", hasNext),
	)

	return data, hasNext, nil
}

func (r *todoRepository) GetAll(userId int) ([]entities.Todo, error) {
	query := "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at FROM todos WHERE user_id = ? AND deleted_at IS NULL ORDER BY id DESC"

	stmt, err := r.db.Prepare(query)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []entities.Todo{}, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(userId)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []entities.Todo{}, err
	}

	defer rows.Close()

	data := []entities.Todo{}

	for rows.Next() {
		var todo entities.Todo

		err = rows.Scan(&todo.ID, &todo.UUID, &todo.Title, &todo.Description, &todo.Status, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt)

		if err != nil {
			return []entities.Todo{}, err
		}

		data = append(data, todo)
	}

	return data, nil
}

func (r *todoRepository) GetByUUID(ctx context.Context, id string, userId int) (entities.Todo, error) {
	query := "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at FROM todos WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRowContext(ctx, query, id)

	var todo entities.Todo

	err := row.Scan(
		&todo.ID,
		&todo.UUID,
		&todo.Title,
		&todo.Description,
		&todo.Status,
		&todo.Completed,
		&todo.UserId,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		slog.Error("Error getting todo by uuid", "error", err)
		return entities.Todo{}, err
	}

	return todo, nil
}

func (r *todoRepository) GetById(ctx context.Context, id string) (entities.Todo, error) {
	query := "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at, deleted_at FROM todos WHERE id = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRowContext(ctx, query, id)

	var todo entities.Todo
	var uuidStr string
	var status int

	err := row.Scan(&todo.ID, &uuidStr, &todo.Title, &todo.Description, &status, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt, &todo.DeletedAt)

	if err != nil {
		return entities.Todo{}, err
	}

	todo.UUID, err = uuid.Parse(uuidStr)
	if err != nil {
		return entities.Todo{}, err
	}

	todo.Status = status

	return todo, nil
}

// TodoRequestInterface defines the interface for todo request parameters
type TodoRequestInterface interface {
	GetTitle() string
	GetDescription() string
	GetStatus() string
	GetCompleted() bool
}

func (r *todoRepository) UpdateByUUID(ctx context.Context, id string, userId int, params interface{}) (entities.Todo, error) {
	// Type assertion to get the TodoRequest
	todoRequest, ok := params.(TodoRequestInterface)
	if !ok {
		return entities.Todo{}, fmt.Errorf("invalid params type")
	}

	oldTodo, err := r.GetByUUID(ctx, id, userId)

	if err != nil {
		return entities.Todo{}, fmt.Errorf("todo with uuid %s not found", id)
	}

	var setParts []string
	var args []interface{}

	// Handle title
	if title := todoRequest.GetTitle(); title != "" {
		setParts = append(setParts, "title = ?")
		args = append(args, title)
	}

	// Handle description
	if description := todoRequest.GetDescription(); description != "" {
		setParts = append(setParts, "description = ?")
		args = append(args, description)
	}

	// Handle status
	if status := todoRequest.GetStatus(); status != "" {
		statusInt, err := StatusToEnum(status)
		if err != nil {
			return entities.Todo{}, err
		}
		setParts = append(setParts, "status = ?")
		args = append(args, statusInt)
	}

	// Handle completed
	setParts = append(setParts, "completed = ?")
	args = append(args, todoRequest.GetCompleted())

	if len(setParts) == 0 {
		return oldTodo, nil
	}

	setParts = append(setParts, "updated_at = ?")

	args = append(args, time.Now())
	args = append(args, id)

	query := fmt.Sprintf("UPDATE todos SET %s WHERE uuid = ? AND deleted_at IS NULL", strings.Join(setParts, ", "))
	stmt, err := r.db.PrepareContext(ctx, query)

	if err != nil {
		return entities.Todo{}, err
	}

	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)

	if err != nil {
		slog.Error("Error updating todo", "error", err)
		return entities.Todo{}, err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return entities.Todo{}, fmt.Errorf("todo with uuid %s not found", id)
	}

	updatedTodo, err := r.GetByUUID(ctx, id, oldTodo.UserId)

	if err != nil {
		return entities.Todo{}, err
	}

	return updatedTodo, nil
}

func (r *todoRepository) DeleteById(id string) error {
	query := "DELETE FROM todos WHERE id = ?"

	result, err := r.db.Exec(query, id)

	if err != nil {
		slog.Error("Error deleting todo", "error", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("todo with id %s not found", id)
	}

	return nil
}

func (r *todoRepository) DeleteByUUID(ctx context.Context, uuid string) error {
	stmt, err := r.db.PrepareContext(ctx, "UPDATE todos SET deleted_at = ? WHERE uuid = ?")

	if err != nil {
		return err
	}

	defer stmt.Close()

	now := time.Now()
	result, err := stmt.ExecContext(ctx, now, uuid)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("todos with uuid %s not found", uuid)
	}

	return nil
}

// StatusToEnum converts string status to enum
func StatusToEnum(status string) (int, error) {
	switch strings.ToLower(status) {
	case "pending":
		return int(entities.TodoStatusPending), nil
	case "in_progress":
		return int(entities.TodoStatusInProgress), nil
	case "in_review":
		return int(entities.TodoStatusInReview), nil
	case "completed":
		return int(entities.TodoStatusCompleted), nil
	default:
		return 0, fmt.Errorf("invalid status: %s. Valid statuses are: pending, in_progress, in_review, completed", status)
	}
}
