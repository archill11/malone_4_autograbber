package pg

import (
	"encoding/json"
	"fmt"
	"myapp/internal/entity"
)

func (s *Database) AddNewGroupLink(title, link string) error {
	q := `INSERT INTO group_link (title, link) 
			VALUES ($1, $2) 
		ON CONFLICT DO NOTHING`
	_, err := s.Exec(q, title, link)
	if err != nil {
		return fmt.Errorf("db: AddNewGroupLink: %w", err)
	}
	return nil
}

func (s *Database) DeleteGroupLink(id int) error {
	q := `DELETE FROM group_link WHERE id = $1`
	_, err := s.Exec(q, id)
	if err != nil {
		return fmt.Errorf("db: DeleteGroupLink: %w", err)
	}
	return nil
}

func (s *Database) UpdateGroupLink(id int, link string) error {
	q := `UPDATE group_link SET link = $1 WHERE id = $2`
	_, err := s.Exec(q, link, id)
	if err != nil {
		return fmt.Errorf("db: UpdateGroupLink: %w", err)
	}
	return nil
}

func (s *Database) GetAllGroupLinks() ([]entity.GroupLink, error) {
	q := `
		SELECT coalesce((
			SELECT json_agg(c)
	  		FROM group_link as c
		), '[]'::json)
	`
	u := make([]entity.GroupLink, 0)
	var data []byte
	err := s.QueryRow(q).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetAllGroupLinks Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetAllGroupLinks Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetGroupLinkById(id int) (entity.GroupLink, error) {
	q := `
		SELECT coalesce((
			SELECT to_json(c)
	  		FROM group_link as c
	  		WHERE id = $1
		), '{}'::json)
	`
	var u entity.GroupLink
	var data []byte
	err := s.QueryRow(q, id).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetGroupLinkById Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetGroupLinkById Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) EditGroupLinkUserCreator(grlLink string, user_creator int) error {
	q := `
		UPDATE group_link SET
			user_creator = $1
		WHERE link = $2
	`
	_, err := s.Exec(q, user_creator, grlLink)
	if err != nil {
		return fmt.Errorf("EditGroupLinkUserCreator: user_creator: %d Link: %d err: %w", user_creator, grlLink, err)

	}
	return nil
}