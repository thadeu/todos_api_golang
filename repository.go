package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        int
	UUID      uuid.UUID
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type UserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
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

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(user User) (User, error) {
	return r.CreateUser(user)
}

func (r *Repository) CreateUser(user User) (User, error) {
	stmt, err := r.db.Prepare("INSERT INTO users (uuid, name, email, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?)")

	if err != nil {
		return User{}, err
	}

	defer stmt.Close()

	uuid := user.UUID.String()

	_, err = stmt.Exec(uuid, user.Name, user.Email, user.CreatedAt, user.UpdatedAt, nil)

	if err != nil {
		return User{}, err
	}

	saved, err := r.GetUserByUUID(uuid)

	if err != nil {
		return User{}, err
	}

	return saved, nil
}

func (r *Repository) GetAllUsers() ([]User, error) {
	rows, err := r.db.Query("SELECT * FROM users WHERE deleted_at IS NULL ORDER BY updated_at DESC")

	if err != nil {
		slog.Error("Error fetching users", "error", err)
		return []User{}, err
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var user User
		var uuidStr string

		err = rows.Scan(&user.ID, &uuidStr, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

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

func (r *Repository) DeleteUser(id string) error {
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

func (r *Repository) GetUserByUUID(uuid string) (User, error) {
	query := "SELECT * FROM users WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(query, uuid)

	var user User

	err := row.Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *Repository) GetUserById(id string) (User, error) {
	query := "SELECT * FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(query, id)

	var user User
	var uuidStr string

	err := row.Scan(&user.ID, &uuidStr, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *Repository) DeleteByUUID(uuid string) error {
	stmt, err := r.db.Prepare("UPDATE users SET deleted_at = ? WHERE uuid = ?")

	if err != nil {
		return err
	}

	defer stmt.Close()

	now := time.Now()
	result, err := stmt.Exec(now, uuid)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("user with uuid %s not found", uuid)
	}

	return nil
}
