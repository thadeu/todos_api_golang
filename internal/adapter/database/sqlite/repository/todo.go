package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"

	"todoapp/internal/adapter/database/sqlite"
	"todoapp/internal/core/domain"
	"todoapp/internal/core/port"
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
	ctx, span := tr.telemetry.StartRepositorySpan(ctx, "GetAllWithCursor", "todo", map[string]interface{}{
		"db.system":         "sqlite",
		"db.table":          "todos",
		"user.id":           userId,
		"pagination.limit":  limit,
		"pagination.cursor": cursor,
	})
	defer span.End()

	// Track operation duration
	startTime := time.Now()

	// Ensure duration is recorded even if function returns early
	defer func() {
		duration := time.Since(startTime)
		span.SetAttributes(map[string]interface{}{
			"operation.duration_ns": duration.Nanoseconds(),
		})
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
			span.SetStatus("error", err.Error())
			span.RecordError(err)
			tr.telemetry.RecordRepositoryOperation(ctx, "GetAllWithCursor", "todo", time.Since(startTime), err)
			return []domain.Todo{}, false, err
		}

		datetime, err := time.Parse(time.RFC3339, datetimeStr)
		if err != nil {
			span.SetStatus("error", err.Error())
			span.RecordError(err)
			tr.telemetry.RecordRepositoryOperation(ctx, "GetAllWithCursor", "todo", time.Since(startTime), err)
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
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "GetAllWithCursor", "todo", time.Since(startTime), err)
		return []domain.Todo{}, false, err
	}

	// Record query details for debugging
	tr.telemetry.RecordRepositoryQuery(ctx, "GetAllWithCursor", "todo", sql, args)

	// Execute database query
	rows, err := tr.db.QueryContext(ctx, sql, args...)
	if err != nil {
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "GetAllWithCursor", "todo", time.Since(startTime), err)
		return []domain.Todo{}, false, err
	}
	defer rows.Close()

	// Scan results
	var todos []domain.Todo
	err = tr.scanner.ScanRowsToSlice(rows, &todos)
	if err != nil {
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "GetAllWithCursor", "todo", time.Since(startTime), err)
		return []domain.Todo{}, false, err
	}

	// Handle pagination
	hasNext := len(todos) == actualLimit
	if hasNext {
		todos = todos[:limit]
	}

	// Update span with operation results
	span.SetAttributes(map[string]interface{}{
		"db.rows_returned": len(todos),
		"db.has_next":      hasNext,
		"db.rows_scanned":  len(todos),
	})

	// Mark operation as successful
	span.SetStatus("ok", "")
	tr.telemetry.RecordRepositoryOperation(ctx, "GetAllWithCursor", "todo", time.Since(startTime), nil)

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
	ctx, span := tr.telemetry.StartRepositorySpan(ctx, "Create", "todo", map[string]interface{}{
		"db.system":    "sqlite",
		"db.table":     "todos",
		"db.operation": "INSERT",
		"todo.uuid":    todo.UUID.String(),
		"user.id":      todo.UserId,
		"todo.title":   todo.Title,
	})
	defer span.End()

	// Track operation duration
	startTime := time.Now()

	// Ensure duration is recorded
	defer func() {
		duration := time.Since(startTime)
		span.SetAttributes(map[string]interface{}{
			"operation.duration_ns": duration.Nanoseconds(),
		})
	}()

	uuid := todo.UUID.String()

	query, args, err := tr.db.QueryBuilder.Insert("todos").
		Columns("uuid", "title", "description", "status", "completed", "user_id", "created_at", "updated_at").
		Values(uuid, todo.Title, todo.Description, todo.Status, todo.Completed, todo.UserId, todo.CreatedAt, todo.UpdatedAt).
		ToSql()

	if err != nil {
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "Create", "todo", time.Since(startTime), err)
		slog.Error("Query build failed", "error", err)
		return domain.Todo{}, err
	}

	// Record the insert query
	tr.telemetry.RecordRepositoryQuery(ctx, "Create", "todo", query, args)

	// Execute insert
	result, err := tr.db.ExecContext(ctx, query, args...)
	if err != nil {
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "Create", "todo", time.Since(startTime), err)
		slog.Error("Insert failed", "error", err, "uuid", uuid)
		return domain.Todo{}, err
	}

	// Get affected rows for verification
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.SetAttributes(map[string]interface{}{"db.result": "unknown"})
	} else {
		span.SetAttributes(map[string]interface{}{"db.rows_affected": rowsAffected})
	}

	// Retrieve the created todo
	saved, err := tr.GetByUUID(ctx, uuid)
	if err != nil {
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "Create", "todo", time.Since(startTime), err)
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
	span.SetStatus("ok", "")
	tr.telemetry.RecordRepositoryOperation(ctx, "Create", "todo", time.Since(startTime), nil)

	return saved, nil
}

func (tr *TodoRepository) UpdateByUUID(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	// Create span using telemetry probe
	ctx, span := tr.telemetry.StartRepositorySpan(ctx, "UpdateByUUID", "todo", map[string]interface{}{
		"db.system":    "sqlite",
		"db.table":     "todos",
		"db.operation": "UPDATE",
		"todo.uuid":    todo.UUID.String(),
		"user.id":      todo.UserId,
	})
	defer span.End()

	// Track operation duration
	startTime := time.Now()

	// Ensure duration is recorded
	defer func() {
		duration := time.Since(startTime)
		span.SetAttributes(map[string]interface{}{
			"operation.duration_ns": duration.Nanoseconds(),
		})
	}()

	// Get current todo for comparison
	oldTodo, err := tr.GetByUUID(ctx, todo.UUID.String())
	if err != nil {
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "UpdateByUUID", "todo", time.Since(startTime), err)
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
	updateAttrs := map[string]interface{}{
		"update.fields_count": len(changes),
	}
	for field, value := range changes {
		updateAttrs[fmt.Sprintf("update.%s", field)] = value
	}
	span.SetAttributes(updateAttrs)

	query, rowArgs, err := tr.db.QueryBuilder.Update("todos").
		SetMap(oldTodo.ToMap()).
		Where(sq.Eq{"uuid": todo.UUID}).
		Where("deleted_at IS NULL").
		ToSql()

	if err != nil {
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "UpdateByUUID", "todo", time.Since(startTime), err)
		return domain.Todo{}, err
	}

	// Record the update query
	tr.telemetry.RecordRepositoryQuery(ctx, "UpdateByUUID", "todo", query, rowArgs)

	// Execute update
	result, err := tr.db.ExecContext(ctx, query, rowArgs...)
	if err != nil {
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "UpdateByUUID", "todo", time.Since(startTime), err)
		return domain.Todo{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.SetAttributes(map[string]interface{}{"db.result": "unknown"})
	} else {
		span.SetAttributes(map[string]interface{}{"db.rows_affected": rowsAffected})
	}

	if rowsAffected == 0 {
		err := fmt.Errorf("no todo updated with uuid %s", todo.UUID)
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "UpdateByUUID", "todo", time.Since(startTime), err)
		return domain.Todo{}, err
	}

	// Get updated todo
	updatedTodo, err := tr.GetByUUID(ctx, todo.UUID.String())
	if err != nil {
		span.SetStatus("error", err.Error())
		span.RecordError(err)
		tr.telemetry.RecordRepositoryOperation(ctx, "UpdateByUUID", "todo", time.Since(startTime), err)
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
	span.SetStatus("ok", "")
	tr.telemetry.RecordRepositoryOperation(ctx, "UpdateByUUID", "todo", time.Since(startTime), nil)

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
