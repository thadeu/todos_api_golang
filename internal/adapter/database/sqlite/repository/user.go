package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"

	"todoapp/internal/adapter/database/sqlite"
	"todoapp/internal/core/domain"
	"todoapp/internal/core/port"
	tel "todoapp/internal/core/telemetry"
)

type UserRepository struct {
	db        *sqlite.DB
	scanner   *sqlite.Scanner
	telemetry port.Telemetry
}

func NewUserRepository(db *sqlite.DB, telemetry port.Telemetry) port.UserRepository {
	if telemetry == nil {
		// Use NoOpProbe if none provided
		telemetry = tel.NewNoOpProbe()
	}

	return &UserRepository{
		db:        db,
		scanner:   sqlite.NewScanner(),
		telemetry: telemetry,
	}
}

func (ur *UserRepository) GetByUUID(ctx context.Context, uid string) (domain.User, error) {
	query := ur.db.QueryBuilder.Select("id", "uuid", "name", "email").
		From("users").
		Where(sq.Eq{"uuid": uid}).
		Limit(1)

	sql, args, err := query.ToSql()

	if err != nil {
		return domain.User{}, err
	}

	var data domain.User

	rows, err := ur.db.QueryContext(ctx, sql, args...)

	if err != nil {
		return domain.User{}, err
	}

	defer rows.Close()

	err = ur.scanner.ScanRowToStruct(rows, &data)

	if err != nil {
		slog.Error("Error getting user by uuid", "error", err)
		return domain.User{}, err
	}

	return data, nil
}

func (ur *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := ur.db.QueryBuilder.Select("*").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1)

	sql, args, err := query.ToSql()

	if err != nil {
		return domain.User{}, err
	}

	var data domain.User

	rows, err := ur.db.QueryContext(ctx, sql, args...)

	if err != nil {
		return domain.User{}, err
	}

	defer rows.Close()

	err = ur.scanner.ScanRowToStruct(rows, &data)

	slog.Info("data", "data", data)

	if err != nil {
		slog.Error("Error getting user by email", "error", err)
		return domain.User{}, err
	}

	return data, nil
}

func (ur *UserRepository) getByUUIDTx(ctx context.Context, tx *sql.Tx, uid string) (domain.User, error) {
	query := ur.db.QueryBuilder.Select("id", "uuid", "name", "email").
		From("users").
		Where(sq.Eq{"uuid": uid}).
		Limit(1)

	sql, args, err := query.ToSql()

	if err != nil {
		return domain.User{}, err
	}

	var data domain.User

	rows, err := tx.QueryContext(ctx, sql, args...)
	if err != nil {
		return domain.User{}, err
	}
	defer rows.Close()

	err = ur.scanner.ScanRowToStruct(rows, &data)

	if err != nil {
		slog.Error("Error getting user by uuid", "error", err)
		return domain.User{}, err
	}

	return data, nil
}

func (ur *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	uuid := user.UUID.String()

	// Use transaction to ensure same connection
	tx, err := ur.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Error starting transaction", "error", err)
		return domain.User{}, err
	}
	defer tx.Rollback()

	query := ur.db.QueryBuilder.Insert("users").
		Columns("uuid", "name", "email", "encrypted_password", "created_at", "updated_at").
		Values(uuid, user.Name, user.Email, user.EncryptedPassword, user.CreatedAt, user.UpdatedAt)

	stmt, args, err := query.ToSql()

	if err != nil {
		slog.Error("Error creating user", "error", err)
		return domain.User{}, err
	}

	_, err = tx.ExecContext(ctx, stmt, args...)

	if err != nil {
		slog.Error("Error creating user", "error", err)
		return domain.User{}, err
	}

	saved, err := ur.getByUUIDTx(ctx, tx, uuid)

	if err != nil {
		return domain.User{}, err
	}

	return saved, tx.Commit()
}

func (ur *UserRepository) DeleteByUUID(ctx context.Context, uuid string) error {
	// Use transaction to ensure same connection
	tx, err := ur.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Error starting transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	query := ur.db.QueryBuilder.Delete("users").
		Where(sq.Eq{"uuid": uuid})

	stmt, args, err := query.ToSql()

	if err != nil {
		slog.Error("Error deleting user", "error", err)
		return err
	}

	result, err := tx.ExecContext(ctx, stmt, args...)

	if err != nil {
		slog.Error("Error deleting user", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("Error getting rows affected", "error", err)
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with uuid %s not found", uuid)
	}

	return tx.Commit()
}
