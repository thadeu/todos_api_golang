package repository

import (
	"context"
	"log/slog"

	sq "github.com/Masterminds/squirrel"

	database "todoapp/internal/adapter/database/postgres"
	domain "todoapp/internal/core/domain"
	port "todoapp/internal/core/port"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) port.UserRepository {
	return &UserRepository{db: db}
}

func (tr *UserRepository) GetByUUID(ctx context.Context, uid string) (domain.User, error) {
	query := tr.db.QueryBuilder.Select("*").
		From("users").
		Where(sq.Eq{"uuid": uid}).
		Limit(1)

	sql, args, err := query.ToSql()

	if err != nil {
		return domain.User{}, err
	}

	var data domain.User

	err = tr.db.QueryRow(ctx, sql, args...).Scan(
		&data.ID,
		&data.UUID,
		&data.Name,
		&data.Email,
		&data.Role,
		&data.CreatedAt,
		&data.UpdatedAt,
	)

	if err != nil {
		slog.Error("Error getting user by uuid", "error", err)
		return domain.User{}, err
	}

	return data, nil
}

func (tr *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := tr.db.QueryBuilder.Select("*").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1)

	sql, args, err := query.ToSql()

	if err != nil {
		return domain.User{}, err
	}

	var data domain.User

	err = tr.db.QueryRow(ctx, sql, args...).Scan(
		&data.ID,
		&data.UUID,
		&data.Name,
		&data.Email,
		&data.Role,
		&data.CreatedAt,
		&data.UpdatedAt,
	)

	if err != nil {
		slog.Error("Error getting user by email", "error", err)
		return domain.User{}, err
	}

	return data, nil
}

func (tr *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	uuid := user.UUID.String()

	query := tr.db.QueryBuilder.Insert("users").
		Columns("uuid", "name", "email", "role", "encrypted_password", "created_at", "updated_at").
		Values(uuid, user.Name, user.Email, user.Role, user.EncryptedPassword, user.CreatedAt, user.UpdatedAt).
		Suffix("RETURNING *")

	stmt, args, err := query.ToSql()

	if err != nil {
		return domain.User{}, err
	}

	err = tr.db.QueryRow(ctx, stmt, args...).Scan(
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.EncryptedPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		slog.Error("Error creating user", "error", err)
		return domain.User{}, err
	}

	saved, err := tr.GetByUUID(ctx, uuid)

	if err != nil {
		return domain.User{}, err
	}

	return saved, nil
}

func (tr *UserRepository) DeleteByUUID(ctx context.Context, uuid string) error {
	stmt, args, err := tr.db.QueryBuilder.Delete("users").
		Where(sq.Eq{"uuid": uuid}).
		ToSql()

	if err != nil {
		return err
	}

	var user domain.User

	err = tr.db.QueryRow(ctx, stmt, args...).Scan(&user.UUID)

	if err != nil {
		return err
	}

	return nil
}
