package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"todoapp/internal/domain/entities"
	"todoapp/internal/domain/repositories"
)

// userRepository implements the UserRepository interface
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) repositories.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user entities.User) (entities.User, error) {
	stmt, err := r.db.PrepareContext(ctx, "INSERT INTO users (uuid, name, email, encrypted_password, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		return entities.User{}, err
	}

	defer stmt.Close()

	uuid := user.UUID.String()

	_, err = stmt.ExecContext(ctx,
		uuid,
		user.Name,
		user.Email,
		user.EncryptedPassword,
		user.CreatedAt,
		user.UpdatedAt,
		nil,
	)

	if err != nil {
		return entities.User{}, err
	}

	saved, err := r.GetUserByUUID(uuid)

	if err != nil {
		return entities.User{}, err
	}

	return saved, nil
}

func (r *userRepository) GetAllUsers() ([]entities.User, error) {
	rows, err := r.db.Query("SELECT id, uuid, name, email, encrypted_password, created_at, updated_at FROM users WHERE deleted_at IS NULL ORDER BY updated_at DESC")

	if err != nil {
		slog.Error("Error fetching users", "error", err)
		return []entities.User{}, err
	}

	defer rows.Close()

	users := []entities.User{}

	for rows.Next() {
		var user entities.User
		var uuidStr string

		err = rows.Scan(&user.ID, &uuidStr, &user.Name, &user.Email, &user.EncryptedPassword, &user.CreatedAt, &user.UpdatedAt)

		if err != nil {
			return []entities.User{}, err
		}

		user.UUID, err = uuid.Parse(uuidStr)

		if err != nil {
			return []entities.User{}, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *userRepository) GetUserByUUID(uuid string) (entities.User, error) {
	query := "SELECT id, uuid, name, email, encrypted_password, created_at, updated_at FROM users WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(query, uuid)

	var user entities.User

	err := row.Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.EncryptedPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserById(id string) (entities.User, error) {
	query := "SELECT id, uuid, name, email, encrypted_password, created_at, updated_at, deleted_at FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(query, id)

	var user entities.User
	var uuidStr string

	err := row.Scan(
		&user.ID,
		&uuidStr,
		&user.Name,
		&user.Email,
		&user.EncryptedPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return entities.User{}, err
	}

	user.UUID, err = uuid.Parse(uuidStr)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	query := "SELECT id, uuid, name, email, encrypted_password, created_at, updated_at FROM users WHERE email = ? LIMIT 1"

	row := r.db.QueryRowContext(ctx, query, email)

	var user entities.User

	scanErr := row.Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.EncryptedPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	slog.Info("User", "user", user)

	if scanErr != nil {
		return user, scanErr
	}

	return user, nil
}

func (r *userRepository) DeleteUser(id string) error {
	query := "DELETE FROM users WHERE id = ?"

	result, err := r.db.Exec(query, id)

	if err != nil {
		log.Println(err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %s not found", id)
	}

	return nil
}

func (r *userRepository) DeleteByUUID(ctx context.Context, uuid string) error {
	stmt, err := r.db.Prepare("UPDATE users SET deleted_at = ? WHERE uuid = ?")

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
		return fmt.Errorf("user with uuid %s not found", uuid)
	}

	return nil
}
