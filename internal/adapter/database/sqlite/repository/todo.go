package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/attribute"

	"todoapp/internal/adapter/database/sqlite"
	"todoapp/internal/core/domain"
	"todoapp/internal/core/port"
	"todoapp/internal/core/util"
	"todoapp/pkg/tracing"
)

type TodoRepository struct {
	db      *sqlite.DB
	scanner *sqlite.Scanner
}

func NewTodoRepository(db *sqlite.DB) port.TodoRepository {
	return &TodoRepository{db: db, scanner: sqlite.NewScanner()}
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

	stmt, err := tr.db.PrepareContext(ctx, query)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []domain.Todo{}, false, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []domain.Todo{}, false, err
	}

	defer rows.Close()

	data := []domain.Todo{}

	for rows.Next() {
		var todo domain.Todo

		err = rows.Scan(&todo.ID, &todo.UUID, &todo.Title, &todo.Description, &todo.Status, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt)

		if err != nil {
			return []domain.Todo{}, false, err
		}

		data = append(data, todo)
	}

	hasNext := len(data) == actualLimit

	if hasNext {
		data = data[:limit]
	}

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

	rows, err := tr.db.QueryContext(ctx, sql, args...)

	if err != nil {
		return domain.Todo{}, err
	}

	defer rows.Close()

	var todo domain.Todo
	err = tr.scanner.ScanRowToStruct(rows, &todo)

	todo.Status, _ = todo.StatusToEnum(todo.StatusOrFallback())

	if err != nil {
		slog.Error("Error getting todo by uuid", "error", err)
		return domain.Todo{}, err
	}

	return todo, nil
}

func (tr *TodoRepository) Create(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	uuid := todo.UUID.String()

	query, args, err := tr.db.QueryBuilder.Insert("todos").
		Columns("uuid", "title", "description", "status", "completed", "user_id", "created_at", "updated_at").
		Values(uuid, todo.Title, todo.Description, todo.Status, todo.Completed, todo.UserId, todo.CreatedAt, todo.UpdatedAt).
		ToSql()

	if err != nil {
		slog.Error("Query build failed", "error", err)
		return domain.Todo{}, err
	}

	_, err = tr.db.ExecContext(ctx, query, args...)

	if err != nil {
		slog.Error("Insert failed", "error", err, "uuid", uuid)
		return domain.Todo{}, err
	}

	saved, err := tr.GetByUUID(ctx, uuid)

	if err != nil {
		slog.Error("GetByUUID failed after insert", "error", err, "uuid", uuid)
		return domain.Todo{}, err
	}

	return saved, nil
}

func (tr *TodoRepository) UpdateByUUID(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	slog.Info("UpdateByUUID", "todo", todo)

	oldTodo, err := tr.GetByUUID(ctx, todo.UUID.String())

	slog.Info("UpdateByUUID", "oldTodo", oldTodo)

	if err != nil {
		slog.Error("UpdateByUUID oldTodo", "error", err)
		return domain.Todo{}, fmt.Errorf("todo with uuid %s not found", todo.UUID)
	}

	if todo.Title != "" {
		oldTodo.Title = todo.Title
	}

	if todo.Description != "" {
		oldTodo.Description = todo.Description
	}

	if todo.Completed == false || todo.Completed == true {
		oldTodo.Completed = todo.Completed
	}

	if todo.Status != 0 {
		oldTodo.Status = todo.Status
	}

	oldTodo.UpdatedAt = time.Now()

	query, rowArgs, err := tr.db.QueryBuilder.Update("todos").
		SetMap(oldTodo.ToMap()).
		Where(sq.Eq{"uuid": todo.UUID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()

	if err != nil {
		return domain.Todo{}, err
	}

	result, err := tr.db.ExecContext(ctx, query, rowArgs...)

	if err != nil {
		slog.Error("Error updating todo", "error", err)
		return domain.Todo{}, err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return domain.Todo{}, fmt.Errorf("no one todo updated with uuid %s", todo.UUID)
	}

	updatedTodo, err := tr.GetByUUID(ctx, todo.UUID.String())

	if err != nil {
		return domain.Todo{}, err
	}

	return updatedTodo, nil
}

func (tr *TodoRepository) DeleteByUUID(ctx context.Context, uuid string) error {
	stmt, err := tr.db.PrepareContext(ctx, "UPDATE todos SET deleted_at = ? WHERE uuid = ?")

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
