package repositories

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	m "todoapp/internal/models"

	"github.com/google/uuid"
)

type TodoRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Completed   bool       `json:"completed,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type TodoResponse struct {
	UUID        uuid.UUID `json:"uuid,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Completed   bool      `json:"completed,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
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

func (r *TodoRepository) Create(todo m.Todo) (m.Todo, error) {
	stmt, err := r.db.Prepare("INSERT INTO todos (uuid, title, description, completed, user_id, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		return m.Todo{}, err
	}

	defer stmt.Close()

	uuid := todo.UUID.String()

	_, err = stmt.Exec(uuid, todo.Title, todo.Description, todo.Completed, todo.UserId, todo.CreatedAt, todo.UpdatedAt, nil)

	if err != nil {
		return m.Todo{}, err
	}

	saved, err := r.GetByUUID(uuid, todo.UserId)

	if err != nil {
		return m.Todo{}, err
	}

	return saved, nil
}

func (r *TodoRepository) GetAll(userId int) ([]m.Todo, error) {
	rows, err := r.db.Query("SELECT * FROM todos WHERE deleted_at IS NULL AND user_id = ? ORDER BY updated_at DESC", userId)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)
		return []m.Todo{}, err
	}

	defer rows.Close()

	data := []m.Todo{}

	for rows.Next() {
		var todo m.Todo
		var uuidStr string

		err = rows.Scan(&todo.ID, &uuidStr, &todo.Title, &todo.Description, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt, &todo.DeletedAt)

		if err != nil {
			return []m.Todo{}, err
		}

		todo.UUID, err = uuid.Parse(uuidStr)

		if err != nil {
			return []m.Todo{}, err
		}

		data = append(data, todo)
	}

	return data, nil
}

func (r *TodoRepository) GetByUUID(uuid string, userId int) (m.Todo, error) {
	query := "SELECT * FROM todos WHERE uuid = ? AND deleted_at IS NULL AND user_id = ? LIMIT 1"

	row := r.db.QueryRow(query, uuid, userId)

	var todo m.Todo

	err := row.Scan(
		&todo.ID,
		&todo.UUID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.UserId,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.DeletedAt,
	)

	if err != nil {
		return m.Todo{}, err
	}

	return todo, nil
}

func (r *TodoRepository) GetById(id string) (m.Todo, error) {
	query := "SELECT * FROM todos WHERE id = ? AND deleted_at IS NULL LIMIT 1"

	row := r.db.QueryRow(query, id)

	var todo m.Todo
	var uuidStr string

	err := row.Scan(&todo.ID, &uuidStr, &todo.Title, &todo.Description, &todo.Completed, &todo.UserId, &todo.CreatedAt, &todo.UpdatedAt, &todo.DeletedAt)

	if err != nil {
		return m.Todo{}, err
	}

	return todo, nil
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

func (r *TodoRepository) DeleteByUUID(uuid string) error {
	stmt, err := r.db.Prepare("UPDATE todos SET deleted_at = ? WHERE uuid = ? AND deleted_at IS NULL")

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
		return fmt.Errorf("todos with uuid %s not found", uuid)
	}

	return nil
}
