package user

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type UserRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type UserResponse struct {
	UUID      string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Email     string     `json:"email,omitempty"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type GetAllUsersResponse struct {
	Size int            `json:"size"`
	Data []UserResponse `json:"data"`
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, user User) (User, error) {
	return r.CreateUser(ctx, user)
}

func (r *UserRepository) CreateUser(ctx context.Context, user User) (User, error) {
	stmt, err := r.db.PrepareContext(ctx, "INSERT INTO users (uuid, name, email, encrypted_password, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		return User{}, err
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
		return User{}, err
	}

	saved, err := r.GetUserByUUID(uuid)

	if err != nil {
		return User{}, err
	}

	return saved, nil
}

func (r *UserRepository) GetAllUsers() ([]User, error) {
	rows, err := r.db.Query("SELECT id, uuid, name, email, encrypted_password, created_at, updated_at FROM users WHERE deleted_at IS NULL ORDER BY updated_at DESC")

	if err != nil {
		slog.Error("Error fetching users", "error", err)
		return []User{}, err
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var user User
		var uuidStr string

		err = rows.Scan(&user.ID, &uuidStr, &user.Name, &user.Email, &user.EncryptedPassword, &user.CreatedAt, &user.UpdatedAt)

		if err != nil {
			return []User{}, err
		}

		user.UUID, err = uuid.Parse(uuidStr)

		if err != nil {
			return []User{}, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) DeleteUser(id string) error {
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

func (r *UserRepository) GetUserByUUID(uuid string) (User, error) {
	query := "SELECT id, uuid, name, email, encrypted_password, created_at, updated_at FROM users WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(query, uuid)

	var user User

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
		return User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetUserById(id string) (User, error) {
	query := "SELECT id, uuid, name, email, encrypted_password, created_at, updated_at, deleted_at FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(query, id)

	var user User
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
		return User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	query := "SELECT id, uuid, name, email, encrypted_password, created_at, updated_at FROM users WHERE email = ? LIMIT 1"

	row := r.db.QueryRowContext(ctx, query, email)

	var user User

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

func (r *UserRepository) DeleteByUUID(ctx context.Context, uuid string) error {
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
