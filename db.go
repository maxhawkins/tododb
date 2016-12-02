package main

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Todo struct {
	ID      int64 `json:"id" db:"todo_id"`
	Version int64 `json:"version"`

	Title string `json:"title"`
	Notes string `json:"notes"`

	DurationMins int `json:"durationMins,omitempty" db:"duration_mins"`

	CompletedAt *time.Time `json:"completedAt,omitempty" db:"completed_at"`
	DueAt       *time.Time `json:"dueAt,omitempty" db:"due_at"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type ListTodosQuery struct {
}

func (l ListTodosQuery) Exec(e sqlx.Queryer) ([]Todo, error) {
	rows, err := e.Queryx(`SELECT * FROM todos`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo

	for rows.Next() {
		var todo Todo
		if err := rows.StructScan(&todo); err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

type GetTodoQuery struct {
	ID int64
}

func (g GetTodoQuery) Exec(e sqlx.Queryer) (Todo, error) {
	var todo Todo

	row := e.QueryRowx(`SELECT * FROM todos WHERE todo_id = $1`, g.ID)
	if err := row.Err(); err != nil {
		return todo, err
	}

	if err := row.StructScan(&todo); err != nil {
		return todo, err
	}

	return todo, nil
}

type UpdateTodoCommand struct {
	Todo Todo
}

func (u *UpdateTodoCommand) Exec(id int64, e sqlx.Execer) error {
	_, err := e.Exec(`
	UPDATE todos
	SET title=$2, notes=$3, duration_mins=$4, created_at=$5, completed_at=$6, due_at=$7
	WHERE todo_id=$1`,
		id,
		u.Todo.Title,
		u.Todo.Notes,
		u.Todo.DurationMins,
		u.Todo.CompletedAt,
		u.Todo.CompletedAt,
		u.Todo.DueAt)
	if err != nil {
		return err
	}

	return nil
}

type SaveTodoCommand struct {
	Todo Todo
}

func (s *SaveTodoCommand) Exec(e sqlx.Queryer) (Todo, error) {
	row := e.QueryRowx(`
	INSERT INTO todos
		(title, notes, duration_mins, created_at, completed_at, due_at)
	VALUES
		($1, $2, $3, NOW(), $4, $5)
	RETURNING todo_id, created_at`,
		s.Todo.Title,
		s.Todo.Notes,
		s.Todo.DurationMins,
		s.Todo.CompletedAt,
		s.Todo.DueAt)

	if err := row.Scan(&s.Todo.ID, &s.Todo.CreatedAt); err != nil {
		return s.Todo, err
	}

	return s.Todo, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS
todos (
	todo_id			SERIAL		PRIMARY KEY,
	version         INTEGER     NOT NULL DEFAULT 0,

	title			TEXT		NOT NULL,
	notes			TEXT		NOT NULL,
	duration_mins	INT			NOT NULL DEFAULT 0,

	completed_at	TIMESTAMP,
	due_at			TIMESTAMP,

	created_at		TIMESTAMP	NOT NULL
);

CREATE TABLE IF NOT EXISTS
lists (
	list_id         SERIAL      PRIMARY KEY,
	data            TEXT        NOT NULL
);

INSERT INTO lists
	(list_id, data)
VALUES
	(0, '[]')
ON CONFLICT DO NOTHING;
`

func SetupTables(e sqlx.Execer) error {
	_, err := e.Exec(schema)
	if err != nil {
		return err
	}

	return nil
}
