package pg

import (
	"encoding/json"
	"fmt"
	"myapp/internal/entity"
)

func (s *Database) GetUserById(id int) (entity.User, error) {
	q := `
		SELECT coalesce((
			SELECT to_json(c)
	  		FROM users as c
	  		WHERE id = $1
		), '{}'::json)
	`
	var u entity.User
	var data []byte
	err := s.QueryRow(q, id).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetUserById Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetUserById Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetUserByUsername(username string) (entity.User, error) {
	q := `
		SELECT coalesce((
			SELECT to_json(c)
	  		FROM users as c
	  		WHERE username = $1
		), '{}'::json)
	`
	var u entity.User
	var data []byte
	err := s.QueryRow(q, username).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetUserByUsername Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetUserByUsername Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) EditAdmin(username string, is_admin int) error {
	q := `UPDATE users SET is_admin = $1 WHERE username = $2`
	_, err := s.Exec(q, is_admin, username)
	if err != nil {
		return fmt.Errorf("EditAdmin: could not save %s, err: %v", username, err)
	}
	return nil
}

func (s *Database) EditAdminById(id int, is_admin int) error {
	q := `UPDATE users SET is_admin = $1 WHERE id = $2`
	_, err := s.Exec(q, is_admin, id)
	if err != nil {
		return fmt.Errorf("EditAdminById: could not save %d, err: %v", id, err)
	}
	return nil
}

func (s *Database) AddNewUser(id int, username, firstname string) error {
	q := `
		INSERT INTO users (id, username, firstname)
			VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`
	_, err := s.Exec(q, id, username, firstname)
	if err != nil {
		return fmt.Errorf("AddNewUser: could not save %d, err: %s", id, err)
	}
	return nil
}
