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
	err := s.db.QueryRow(q, id).Scan(&u.Id, &u.Username, &u.Firstname, &u.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, repository.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("GetUserById: %w", err)
	}
	return u, nil
}

func (s *Database) GetUserByUsername(username string) (entity.User, error) {
	var u entity.User
	q := `SELECT id, username, firstname, is_admin FROM users WHERE username = $1`
	err := s.db.QueryRow(q, username).Scan(&u.Id, &u.Username, &u.Firstname, &u.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, repository.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("GetUserByUsername: %w", err)
	}
	return u, nil
}

func (s *Database) EditAdmin(username string, is_admin int) error {
	q := `UPDATE users SET is_admin = $1 WHERE username = $2`
	_, err := s.db.Exec(q, is_admin, username)
	if err != nil {
		s.l.Err("Postgres: could not update users table")
		return err
	} else {
		s.l.Info("Postgres: update users table")
		return nil
	}
}

func (s *Database) AddNewUser(id int, username, firstname string) error {
	q := `INSERT INTO users
		(id, username, firstname)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`
	_, err := s.db.Exec(q, id, username, firstname)
	if err != nil {
		s.l.Err("Postgres: could not save the user %d: %s", id, err)
		return err
	} else {
		s.l.Info("Postgres: save %d user", id)
	}
	return nil
}
