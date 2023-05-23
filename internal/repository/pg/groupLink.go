package pg

import (
	"database/sql"
	"errors"
	"fmt"
	"myapp/internal/entity"
	"myapp/internal/repository"
)

func (s *Database) AddNewGroupLink(title, link string) error {
	q := `INSERT INTO group_link (title, link) 
		VALUES ($1, $2) 
		ON CONFLICT DO NOTHING`
	_, err := s.db.Exec(q, title, link)
	if err != nil {
		s.l.Err("Postgres: could not save the group-link")
		err = fmt.Errorf("AddNewGroupLink: %w", err)
		return err
	} else {
		s.l.Info("Postgres: save group-link")
	}
	return nil
}

func (s *Database) DeleteGroupLink(id int) error {
	q := `DELETE FROM group_link WHERE id = $1`
	_, err := s.db.Exec(q, id)
	if err != nil {
		s.l.Err("Postgres: ERR: could not delete the group-link")
		err = fmt.Errorf("DeleteGroupLink: %w", err)
		return err
	} else {
		s.l.Info("Postgres: delete group-link")
	}
	return nil
}

func (s *Database) UpdateGroupLink(id int, link string) error {
	q := `UPDATE group_link SET link = $1 WHERE id = $2`
	_, err := s.db.Exec(q, link, id)
	if err != nil {
		s.l.Err("Postgres: could not change group-link", id, link)
		err = fmt.Errorf("UpdateGroupLink: %w", err)
	} else {
		s.l.Info("Postgres: change group link", id, link)
	}
	return err
}

func (s *Database) GetAllGroupLinks() ([]entity.GroupLink, error) {
	bots := make([]entity.GroupLink, 0)
	q := `SELECT 
			id,
			title,
			link
		FROM group_link`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("GetAllGroupLinks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var b entity.GroupLink
		if err := rows.Scan(&b.Id, &b.Title, &b.Link); err != nil {
			return nil, fmt.Errorf("GetAllGroupLinks (2): %w", err)
		}
		bots = append(bots, b)
	}
	return bots, nil
}

func (s *Database) GetGroupLinkById(id int) (entity.GroupLink, error) {
	var b entity.GroupLink
	q := `SELECT 
			id,
			title,
			link
		FROM group_link
		WHERE id = $1`
	err := s.db.QueryRow(q, id).Scan(
		&b.Id,
		&b.Title,
		&b.Link,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return b, repository.ErrNotFound
		}
		return b, fmt.Errorf("GetGroupLinkById: %w", err)
	}
	return b, nil
}
