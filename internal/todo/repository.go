package todo

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	c "todoapp/pkg/db/cursor"
	. "todoapp/pkg/tracing"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

type TodoStatus int

const (
	TodoStatusPending TodoStatus = iota
	TodoStatusInProgress
	TodoStatusInReview
	TodoStatusCompleted
)

func (t TodoStatus) String() string {
	return []string{"pending", "in_progress", "in_review", "completed"}[t]
}

func StatusToEnum(status string) (int, error) {
	switch strings.ToLower(status) {
	case "pending":
		return int(TodoStatusPending), nil
	case "in_progress":
		return int(TodoStatusInProgress), nil
	case "in_review":
		return int(TodoStatusInReview), nil
	case "completed":
		return int(TodoStatusCompleted), nil
	default:
		return 0, fmt.Errorf("invalid status: %s. Valid statuses are: pending, in_progress, in_review, completed", status)
	}
}

type TodoRequest struct {
	Title       string     `json:"title" validate:"required,min=3,max=255"`
	Description string     `json:"description,omitempty" validate:"max=1000"`
	Status      string     `json:"status,omitempty"`
	Completed   bool       `json:"completed,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type TodoResponse struct {
	UUID        uuid.UUID `json:"uuid"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GetAllTodosResponse struct {
	Size int            `json:"size"`
	Data []TodoResponse `json:"data"`
}

type TodoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) Create(ctx context.Context, todo Todo) (Todo, error) {
	stmt, err := r.db.PrepareContext(ctx, "INSERT INTO todos (uuid, title, description, status, completed, user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		return Todo{}, err
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
		return Todo{}, err
	}

	saved, err := r.GetByUUID(ctx, uuid, todo.UserId)

	if err != nil {
		return Todo{}, err
	}

	return saved, nil
}

func (r *TodoRepository) GetAllWithCursor(ctx context.Context, userId int, limit int, cursor string) ([]Todo, bool, error) {
	// Criar span para operação de banco de dados
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
			return []Todo{}, false, err
		}

		// Parse the datetime string to ensure proper comparison
		datetime, err := time.Parse(time.RFC3339, datetimeStr)

		if err != nil {
			slog.Error("Error parsing cursor datetime", "error", err, "datetime", datetimeStr)
			return []Todo{}, false, err
		}

		query = "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at FROM todos WHERE user_id = ? AND (created_at < ? OR (created_at = ? AND id < ?)) AND deleted_at IS NULL ORDER BY created_at DESC, id DESC LIMIT ?"
		args = []interface{}{userId, datetime, datetime, id, actualLimit}
	}

	stmt, err := r.db.PrepareContext(ctx, query)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []Todo{}, false, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []Todo{}, false, err
	}

	defer rows.Close()

	data := []Todo{}

	for rows.Next() {
		var todo Todo

		err = rows.Scan(&todo.ID, &todo.UUID, &todo.Title, &todo.Description, &todo.Status, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt)

		if err != nil {
			return []Todo{}, false, err
		}

		data = append(data, todo)
	}

	// Verificar se tem próxima página
	hasNext := len(data) == actualLimit

	// Se tem mais dados que o limit, remover o item extra
	if hasNext {
		data = data[:limit]
	}

	// Adicionar atributos de sucesso
	span.SetAttributes(
		attribute.Int("db.rows_returned", len(data)),
		attribute.Bool("db.has_next", hasNext),
	)

	return data, hasNext, nil
}

func (r *TodoRepository) GetAll(userId int) ([]Todo, error) {
	query := "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at FROM todos WHERE user_id = ? AND deleted_at IS NULL ORDER BY id DESC"

	stmt, err := r.db.Prepare(query)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []Todo{}, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(userId)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []Todo{}, err
	}

	defer rows.Close()

	data := []Todo{}

	for rows.Next() {
		var todo Todo

		err = rows.Scan(&todo.ID, &todo.UUID, &todo.Title, &todo.Description, &todo.Status, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt)

		if err != nil {
			return []Todo{}, err
		}

		if err != nil {
			return []Todo{}, err
		}

		data = append(data, todo)
	}

	return data, nil
}

func (r *TodoRepository) GetByUUID(ctx context.Context, id string, userId int) (Todo, error) {
	query := "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at FROM todos WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRowContext(ctx, query, id)

	var todo Todo

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
		return Todo{}, err
	}

	return todo, nil
}

func (r *TodoRepository) GetById(ctx context.Context, id string) (Todo, error) {
	query := "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at, deleted_at FROM todos WHERE id = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRowContext(ctx, query, id)

	var todo Todo
	var uuidStr string
	var status int

	err := row.Scan(&todo.ID, &uuidStr, &todo.Title, &todo.Description, &status, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt, &todo.DeletedAt)

	if err != nil {
		return Todo{}, err
	}

	return todo, nil
}

func (r *TodoRepository) UpdateByUUID(ctx context.Context, id string, userId int, params TodoRequest) (Todo, error) {
	oldTodo, err := r.GetByUUID(ctx, id, userId)

	if err != nil {
		return Todo{}, fmt.Errorf("todo with uuid %s not found", id)
	}

	var setParts []string
	var args []interface{}

	v := reflect.ValueOf(params)
	t := reflect.TypeOf(params)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.IsZero() || fieldType.Name == "CreatedAt" || fieldType.Name == "UpdatedAt" || fieldType.Name == "DeletedAt" {
			continue
		}

		columnName := strings.ToLower(fieldType.Name)

		if fieldType.Name == "Status" && field.Kind() == reflect.String {
			statusStr := field.String()

			if statusStr != "" {
				statusInt, err := StatusToEnum(statusStr)

				if err != nil {
					return Todo{}, err
				}

				setParts = append(setParts, fmt.Sprintf("%s = ?", columnName))
				args = append(args, statusInt)

				continue
			}
		}

		setParts = append(setParts, fmt.Sprintf("%s = ?", columnName))
		args = append(args, field.Interface())
	}

	if len(setParts) == 0 {
		return oldTodo, nil
	}

	setParts = append(setParts, "updated_at = ?")

	args = append(args, time.Now())
	args = append(args, id)

	query := fmt.Sprintf("UPDATE todos SET %s WHERE uuid = ? AND deleted_at IS NULL", strings.Join(setParts, ", "))
	stmt, err := r.db.PrepareContext(ctx, query)

	if err != nil {
		return Todo{}, err
	}

	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)

	if err != nil {
		slog.Error("Error updating todo", "error", err)
		return Todo{}, err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return Todo{}, fmt.Errorf("todo with uuid %s not found", id)
	}

	updatedTodo, err := r.GetByUUID(ctx, id, oldTodo.UserId)

	if err != nil {
		return Todo{}, err
	}

	return updatedTodo, nil
}

func (r *TodoRepository) DeleteById(id string) error {
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

func (r *TodoRepository) DeleteByUUID(ctx context.Context, uuid string) error {
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
