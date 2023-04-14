package pg

import (
	"database/sql"
	"errors"
	"myapp/internal/entity"
	"myapp/internal/repository"
)

func (s *Database) AddNewGroupLink(title, link string) error {
	q := `INSERT INTO group_link (title, link) 
		VALUES ($1, $2) 
		ON CONFLICT DO NOTHING`
	_, err := s.db.Exec(q, title, link)
	if err != nil {
		s.l.Err("Postgres: could not save the group link")
		return err
	} else {
		s.l.Info("Postgres: save group link")
	}
	return nil
}

func (s *Database) DeleteGroupLink(id int) error {
	q := `DELETE FROM group_link WHERE id = $1`
	_, err := s.db.Exec(q, id)
	if err != nil {
		s.l.Err("Postgres: ERR: could not delete the group link")
		return err
	} else {
		s.l.Info("Postgres: delete group link")
	}
	return nil
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
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var b entity.GroupLink
		if err := rows.Scan(&b.Id, &b.Title, &b.Link); err != nil {
			return nil, err
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
		return b, err
	}
	return b, nil
}