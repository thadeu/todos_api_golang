package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/attribute"

	"todoapp/internal/adapter/database/postgres"
	"todoapp/internal/core/domain"
	"todoapp/internal/core/port"
	"todoapp/internal/core/util"
	"todoapp/pkg/tracing"
)

type TodoRepository struct {
	db      *postgres.DB
	scanner *postgres.Scanner
}

func NewTodoRepository(db *postgres.DB) port.TodoRepository {
	return &TodoRepository{db: db, scanner: postgres.NewScanner()}
}

func (tr *TodoRepository) GetAllWithCursor(ctx context.Context, userId int, limit int, cursor string) ([]domain.Todo, bool, error) {
	ctx, span := tracing.CreateChildSpan(ctx, "db.todo.GetAllWithCursor", []attribute.KeyValue{
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
		datetimeStr, id, err := util.DecodeCursor(cursor)

		if err != nil {
			tracing.AddSpanError(span, err)
			slog.Error("Error decoding cursor", "error", err)

			return []domain.Todo{}, false, err
		}

		datetime, err := time.Parse(time.RFC3339, datetimeStr)

		if err != nil {
			slog.Error("Error parsing cursor datetime", "error", err, "datetime", datetimeStr)
			return []domain.Todo{}, false, err
		}

		query = "SELECT id, uuid, title, description, status, completed, user_id, created_at, updated_at FROM todos WHERE user_id = ? AND (created_at < ? OR (created_at = ? AND id < ?)) AND deleted_at IS NULL ORDER BY created_at DESC, id DESC LIMIT ?"
		args = []interface{}{userId, datetime, datetime, id, actualLimit}
	}

	rows, err := tr.db.Query(ctx, query, args...)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []domain.Todo{}, false, err
	}

	data := []domain.Todo{}

	for rows.Next() {
		var todo domain.Todo

		err = rows.Scan(&todo.ID, &todo.UUID, &todo.Title, &todo.Description, &todo.Status, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt)

		if err != nil {
			return []domain.Todo{}, false, err
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

func (tr *TodoRepository) GetByUUID(ctx context.Context, uid string) (domain.Todo, error) {
	query := tr.db.QueryBuilder.Select("*").
		From("todos").
		Where(sq.Eq{"uuid": uid}).
		Where(sq.Eq{"deleted_at": nil}).
		Limit(1)

	sql, args, err := query.ToSql()

	if err != nil {
		return domain.Todo{}, err
	}

	rows, err := tr.db.Query(ctx, sql, args...)

	if err != nil {
		return domain.Todo{}, err
	}

	defer rows.Close()

	var todo domain.Todo
	err = tr.scanner.ScanRowsToSlice(rows, &todo)

	todo.Status, _ = todo.StatusToEnum(todo.StatusOrFallback())

	if err != nil {
		slog.Error("Error getting todo by uuid", "error", err)
		return domain.Todo{}, err
	}

	return todo, nil
}

func (tr *TodoRepository) Create(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	uuid := todo.UUID.String()

	query := tr.db.QueryBuilder.Insert("todos").
		Columns("uuid", "title", "description", "status", "completed", "user_id", "created_at", "updated_at").
		Values(uuid, todo.Title, todo.Description, todo.Completed, todo.UserId, todo.CreatedAt, todo.UpdatedAt).
		Suffix("RETURNING *")

	stmt, args, err := query.ToSql()

	if err != nil {
		return domain.Todo{}, err
	}

	err = tr.db.QueryRow(ctx, stmt, args...).Scan(
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
		slog.Error("Error creating todo", "error", err)
		return domain.Todo{}, err
	}

	saved, err := tr.GetByUUID(ctx, uuid)

	if err != nil {
		return domain.Todo{}, err
	}

	return saved, nil
}

func (tr *TodoRepository) UpdateByUUID(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	oldTodo, err := tr.GetByUUID(ctx, todo.UUID.String())

	if err != nil {
		return domain.Todo{}, fmt.Errorf("todo with uuid %s not found", todo.UUID)
	}

	var setParts []string
	var args []interface{}

	if title := todo.Title; title != "" {
		setParts = append(setParts, "title = ?")
		args = append(args, title)
	}

	if description := todo.Description; description != "" {
		setParts = append(setParts, "description = ?")
		args = append(args, description)
	}

	if status := todo.Status; status != 0 {
		statusInt := todo.StatusOrFallback("")

		if err != nil {
			return domain.Todo{}, err
		}

		setParts = append(setParts, "status = ?")
		args = append(args, statusInt)
	}

	// Handle completed
	setParts = append(setParts, "completed = ?")
	args = append(args, todo.Completed)

	if len(setParts) == 0 {
		return oldTodo, nil
	}

	setParts = append(setParts, "updated_at = ?")

	args = append(args, time.Now())
	args = append(args, todo.ID)

	query := tr.db.QueryBuilder.Update("todos").
		SetMap(map[string]interface{}{
			"title":       todo.Title,
			"description": todo.Description,
			"status":      todo.Status,
			"completed":   todo.Completed,
			"updated_at":  time.Now(),
		}).
		Where(sq.Eq{"uuid": todo.UUID}).
		Where(sq.Eq{"deleted_at": nil})

	stmt, args, err := query.ToSql()

	if err != nil {
		return domain.Todo{}, err
	}

	err = tr.db.QueryRow(ctx, stmt, args...).Scan(
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
		slog.Error("Error updating todo", "error", err)
		return domain.Todo{}, err
	}

	updatedTodo, err := tr.GetByUUID(ctx, todo.UUID.String())

	if err != nil {
		return domain.Todo{}, err
	}

	return updatedTodo, nil
}

func (tr *TodoRepository) DeleteByUUID(ctx context.Context, uuid string) error {
	query := tr.db.QueryBuilder.Update("todos").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"uuid": uuid})

	stmt, args, err := query.ToSql()

	if err != nil {
		return err
	}

	var todo domain.Todo
	err = tr.db.QueryRow(ctx, stmt, args...).Scan(&todo.UUID)

	if err != nil {
		return fmt.Errorf("todos with uuid %s not found", uuid)
	}

	return nil
}
