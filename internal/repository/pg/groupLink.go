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
		return fmt.Errorf("db: AddNewGroupLink: %w", err)
	}
	return nil
}

func (s *Database) DeleteGroupLink(id int) error {
	q := `DELETE FROM group_link WHERE id = $1`
	_, err := s.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("db: DeleteGroupLink: %w", err)
	}
	return nil
}

func (s *Database) UpdateGroupLink(id int, link string) error {
	q := `UPDATE group_link SET link = $1 WHERE id = $2`
	_, err := s.db.Exec(q, link, id)
	if err != nil {
		return fmt.Errorf("db: UpdateGroupLink: %w", err)
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
		return nil, fmt.Errorf("db: GetAllGroupLinks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var b entity.GroupLink
		if err := rows.Scan(&b.Id, &b.Title, &b.Link); err != nil {
			return nil, fmt.Errorf("db: GetAllGroupLinks (2): %w", err)
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
		return b, fmt.Errorf("db: GetGroupLinkById: %w", err)
	}
	return b, nil
}
