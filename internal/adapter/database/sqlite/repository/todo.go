package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"todoapp/internal/adapter/database/sqlite"
	"todoapp/internal/core/domain"
	"todoapp/internal/core/port"
	"todoapp/internal/core/telemetry"
	tel "todoapp/internal/core/telemetry"
	"todoapp/internal/core/util"
)

type TodoRepository struct {
	db        *sqlite.DB
	scanner   *sqlite.Scanner
	telemetry port.Telemetry
}

func NewTodoRepository(db *sqlite.DB, telemetry port.Telemetry) port.TodoRepository {
	if telemetry == nil {
		telemetry = tel.NewNoOpProbe()
	}

	return &TodoRepository{
		db:        db,
		scanner:   sqlite.NewScanner(),
		telemetry: telemetry,
	}
}

func (tr *TodoRepository) GetAllWithCursor(ctx context.Context, userId int, limit int, cursor string) ([]domain.Todo, bool, error) {
	// Create span using telemetry probe
	ctx, span := tr.telemetry.StartRepositorySpan(ctx, "GetAllWithCursor", "todo", []attribute.KeyValue{
		attribute.String("db.system", "sqlite"),
		attribute.String("db.table", "todos"),
		attribute.Int("user.id", userId),
		attribute.Int("pagination.limit", limit),
		attribute.String("pagination.cursor", cursor),
	})
	defer span.End()

	// Start telemetry operation for metrics and logging
	startTime := time.Now()
	operation := telemetry.StartOperation(tr.telemetry, ctx, "GetAllWithCursor", "todo")

	// Ensure operation is recorded even if function returns early
	defer func() {
		duration := time.Since(startTime).Nanoseconds()
		span.SetAttributes(attribute.Int64("operation.duration_ns", duration))
	}()

	actualLimit := limit + 1

	query := tr.db.QueryBuilder.Select("*").
		From("todos").
		Where(sq.Eq{"user_id": userId}).
		Where("deleted_at IS NULL").
		OrderBy("created_at DESC, id DESC").
		Limit(uint64(actualLimit))

	if cursor != "" {
		datetimeStr, id, err := util.DecodeCursor(cursor)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			operation.End(err)
			return []domain.Todo{}, false, err
		}

		datetime, err := time.Parse(time.RFC3339, datetimeStr)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			operation.End(err)
			return []domain.Todo{}, false, err
		}

		query = query.Where(sq.Or{
			sq.Lt{"created_at": datetime},
			sq.And{
				sq.Eq{"created_at": datetime},
				sq.Lt{"id": id},
			},
		})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		return []domain.Todo{}, false, err
	}

	// Record query details for debugging
	tr.telemetry.RecordRepositoryQuery(ctx, "GetAllWithCursor", "todo", sql, args)

	// Execute database query
	rows, err := tr.db.QueryContext(ctx, sql, args...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		return []domain.Todo{}, false, err
	}
	defer rows.Close()

	// Scan results
	var todos []domain.Todo
	err = tr.scanner.ScanRowsToSlice(rows, &todos)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		return []domain.Todo{}, false, err
	}

	// Handle pagination
	hasNext := len(todos) == actualLimit
	if hasNext {
		todos = todos[:limit]
	}

	// Update span with operation results
	span.SetAttributes(
		attribute.Int("db.rows_returned", len(todos)),
		attribute.Bool("db.has_next", hasNext),
		attribute.Int("db.rows_scanned", len(todos)),
	)

	// Mark operation as successful
	operation.End(nil)

	return todos, hasNext, nil
}

func (tr *TodoRepository) GetByUUID(ctx context.Context, uid string) (domain.Todo, error) {
	query := tr.db.QueryBuilder.Select("*").
		From("todos").
		Where(sq.Eq{"uuid": uid}).
		Where("deleted_at IS NULL").
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
	// Create span using telemetry probe
	ctx, span := tr.telemetry.StartRepositorySpan(ctx, "Create", "todo", []attribute.KeyValue{
		attribute.String("db.system", "sqlite"),
		attribute.String("db.table", "todos"),
		attribute.String("db.operation", "INSERT"),
		attribute.String("todo.uuid", todo.UUID.String()),
		attribute.Int("user.id", todo.UserId),
		attribute.String("todo.title", todo.Title),
	})
	defer span.End()

	// Start telemetry operation
	startTime := time.Now()
	operation := telemetry.StartOperation(tr.telemetry, ctx, "Create", "todo")

	// Ensure operation is recorded
	defer func() {
		duration := time.Since(startTime).Nanoseconds()
		span.SetAttributes(attribute.Int64("operation.duration_ns", duration))
	}()

	uuid := todo.UUID.String()

	query, args, err := tr.db.QueryBuilder.Insert("todos").
		Columns("uuid", "title", "description", "status", "completed", "user_id", "created_at", "updated_at").
		Values(uuid, todo.Title, todo.Description, todo.Status, todo.Completed, todo.UserId, todo.CreatedAt, todo.UpdatedAt).
		ToSql()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		slog.Error("Query build failed", "error", err)
		return domain.Todo{}, err
	}

	// Record the insert query
	tr.telemetry.RecordRepositoryQuery(ctx, "Create", "todo", query, args)

	// Execute insert
	result, err := tr.db.ExecContext(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		slog.Error("Insert failed", "error", err, "uuid", uuid)
		return domain.Todo{}, err
	}

	// Get affected rows for verification
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.SetAttributes(attribute.String("db.result", "unknown"))
	} else {
		span.SetAttributes(attribute.Int64("db.rows_affected", rowsAffected))
	}

	// Retrieve the created todo
	saved, err := tr.GetByUUID(ctx, uuid)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		slog.Error("GetByUUID failed after insert", "error", err, "uuid", uuid)
		return domain.Todo{}, err
	}

	// Record business event
	tr.telemetry.RecordBusinessEvent(ctx, "created", "todo", saved.UUID.String(), saved.UserId, map[string]interface{}{
		"title":      saved.Title,
		"status":     saved.StatusOrFallback(),
		"created_at": saved.CreatedAt,
	})

	// Mark operation as successful
	operation.End(nil)

	return saved, nil
}

func (tr *TodoRepository) UpdateByUUID(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	// Create span using telemetry probe
	ctx, span := tr.telemetry.StartRepositorySpan(ctx, "UpdateByUUID", "todo", []attribute.KeyValue{
		attribute.String("db.system", "sqlite"),
		attribute.String("db.table", "todos"),
		attribute.String("db.operation", "UPDATE"),
		attribute.String("todo.uuid", todo.UUID.String()),
		attribute.Int("user.id", todo.UserId),
	})
	defer span.End()

	// Start telemetry operation
	startTime := time.Now()
	operation := telemetry.StartOperation(tr.telemetry, ctx, "UpdateByUUID", "todo")

	// Ensure operation is recorded
	defer func() {
		duration := time.Since(startTime).Nanoseconds()
		span.SetAttributes(attribute.Int64("operation.duration_ns", duration))
	}()

	// Get current todo for comparison
	oldTodo, err := tr.GetByUUID(ctx, todo.UUID.String())
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		return domain.Todo{}, fmt.Errorf("todo with uuid %s not found", todo.UUID)
	}

	// Track what fields are being updated
	changes := make(map[string]interface{})

	// Apply updates
	if todo.Title != "" && todo.Title != oldTodo.Title {
		oldTodo.Title = todo.Title
		changes["title"] = todo.Title
	}

	if todo.Description != "" && todo.Description != oldTodo.Description {
		oldTodo.Description = todo.Description
		changes["description"] = todo.Description
	}

	if (todo.Completed == false || todo.Completed == true) && todo.Completed != oldTodo.Completed {
		oldTodo.Completed = todo.Completed
		changes["completed"] = todo.Completed
	}

	if todo.Status != 0 && todo.Status != oldTodo.Status {
		oldTodo.Status = todo.Status
		changes["status"] = todo.Status
	}

	oldTodo.UpdatedAt = time.Now()

	// Add changes to span
	for field, value := range changes {
		span.SetAttributes(attribute.String(fmt.Sprintf("update.%s", field), fmt.Sprintf("%v", value)))
	}
	span.SetAttributes(attribute.Int("update.fields_count", len(changes)))

	query, rowArgs, err := tr.db.QueryBuilder.Update("todos").
		SetMap(oldTodo.ToMap()).
		Where(sq.Eq{"uuid": todo.UUID}).
		Where("deleted_at IS NULL").
		ToSql()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		return domain.Todo{}, err
	}

	// Record the update query
	tr.telemetry.RecordRepositoryQuery(ctx, "UpdateByUUID", "todo", query, rowArgs)

	// Execute update
	result, err := tr.db.ExecContext(ctx, query, rowArgs...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		return domain.Todo{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.SetAttributes(attribute.String("db.result", "unknown"))
	} else {
		span.SetAttributes(attribute.Int64("db.rows_affected", rowsAffected))
	}

	if rowsAffected == 0 {
		err := fmt.Errorf("no todo updated with uuid %s", todo.UUID)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		return domain.Todo{}, err
	}

	// Get updated todo
	updatedTodo, err := tr.GetByUUID(ctx, todo.UUID.String())
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		operation.End(err)
		return domain.Todo{}, err
	}

	// Record business event if there were changes
	if len(changes) > 0 {
		tr.telemetry.RecordBusinessEvent(ctx, "updated", "todo", updatedTodo.UUID.String(), updatedTodo.UserId, map[string]interface{}{
			"changes":    changes,
			"updated_at": updatedTodo.UpdatedAt,
		})
	}

	// Mark operation as successful
	operation.End(nil)

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
