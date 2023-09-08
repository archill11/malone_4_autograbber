package pg

import (
	"database/sql"
	"errors"
	"fmt"
	"myapp/internal/entity"
	"myapp/internal/repository"
)

func (s *Database) GetUserById(id int) (entity.User, error) {
	var u entity.User
	q := `SELECT id, username, firstname, is_admin FROM users WHERE id = $1`
	err := s.QueryRow(q, id).Scan(&u.Id, &u.Username, &u.Firstname, &u.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, repository.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("db: GetUserById: %w", err)
	}
	return u, nil
}

func (s *Database) GetUserByUsername(username string) (entity.User, error) {
	var u entity.User
	q := `SELECT id, username, firstname, is_admin FROM users WHERE username = $1`
	err := s.QueryRow(q, username).Scan(&u.Id, &u.Username, &u.Firstname, &u.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, repository.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("db: GetUserByUsername: %w", err)
	}
	return u, nil
}

func (s *Database) EditAdmin(username string, is_admin int) error {
	q := `UPDATE users SET is_admin = $1 WHERE username = $2`
	_, err := s.Exec(q, is_admin, username)
	if err != nil {
		return fmt.Errorf("db: EditAdmin: %w", err)
	}
	return nil
}

func (s *Database) AddNewUser(id int, username, firstname string) error {
	q := `INSERT INTO users
		(id, username, firstname)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`
	_, err := s.Exec(q, id, username, firstname)
	if err != nil {
		return fmt.Errorf("db: AddNewUser: %w", err)
	}
	return nil
}
