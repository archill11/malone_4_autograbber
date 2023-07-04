package pg

import (
	"database/sql"
	"errors"
	"fmt"
	"myapp/internal/entity"
	"myapp/internal/repository"
)

func (s *Database) AddNewBot(id int, username, firstname, token string, idDonor int) error {
	e := entity.NewBot(id, username, firstname, token, idDonor)
	q := `INSERT INTO bots (id, username, first_name, token, is_donor) 
		VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT DO NOTHING`
	_, err := s.db.Exec(q, e.Id, e.Username, e.Firstname, e.Token, e.IsDonor)
	if err != nil {
		return fmt.Errorf("db: AddNewBot: %w", err)
	}
	return nil
}

func (s *Database) DeleteBot(id int) error {
	q := `DELETE FROM bots WHERE id = $1`
	_, err := s.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("db: DeleteBot: %w", err)
	}
	return nil
}

func (s *Database) GetBotByChannelId(channelId int) (entity.Bot, error) {
	var b entity.Bot
	q := `SELECT 
			id,
			username,
			first_name,
			token,
			is_donor,
			ch_id,
			ch_link,
			group_link_id
		FROM bots
		WHERE ch_id = $1`
	err := s.db.QueryRow(q, channelId).Scan(
		&b.Id,
		&b.Username,
		&b.Firstname,
		&b.Token,
		&b.IsDonor,
		&b.ChId,
		&b.ChLink,
		&b.GroupLinkId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return b, fmt.Errorf("db: GetBotByChannelId: channelId: %d ErrNotFound", channelId)
		}
		return b, fmt.Errorf("db: GetBotByChannelId: channelId: %d err: %w", channelId, err)
	}
	return b, nil
}

func (s *Database) GetBotsByGrouLinkId(groupLinkId int) ([]entity.Bot, error) {
	bots := make([]entity.Bot, 0)
	q := `SELECT 
			id,
			username,
			first_name,
			token,
			is_donor,
			ch_id,
			ch_link,
			group_link_id
		FROM bots
		WHERE group_link_id = $1`
	rows, err := s.db.Query(q, groupLinkId)
	if err != nil {
		return nil, fmt.Errorf("db: GetBotsByGrouLinkId: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var b entity.Bot
		if err := rows.Scan(
			&b.Id,
			&b.Username,
			&b.Firstname,
			&b.Token,
			&b.IsDonor,
			&b.ChId,
			&b.ChLink,
			&b.GroupLinkId,
		); err != nil {
			return nil, fmt.Errorf("db: GetBotsByGrouLinkId (2): %w", err)
		}
		bots = append(bots, b)
	}
	return bots, nil
}

func (s *Database) GetAllBots() ([]entity.Bot, error) {
	bots := make([]entity.Bot, 0)
	q := `SELECT 
			id,
			token,
			username,
			first_name,
			is_donor,
			ch_id,
			ch_link,
			group_link_id
		FROM bots`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("GetAllBots: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var b entity.Bot
		if err := rows.Scan(
			&b.Id,
			&b.Token,
			&b.Username,
			&b.Firstname,
			&b.IsDonor,
			&b.ChId,
			&b.ChLink,
			&b.GroupLinkId,
		); err != nil {
			return nil, fmt.Errorf("db: GetAllBots (2): %w", err)
		}
		bots = append(bots, b)
	}
	return bots, nil
}

func (s *Database) GetAllVampBots() ([]entity.Bot, error) {
	bots := make([]entity.Bot, 0)
	q := `SELECT 
			id,
			token,
			username,
			first_name,
			is_donor,
			ch_id,
			ch_link,
			group_link_id
		FROM bots
		WHERE is_donor = 0`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("db: GetAllVampBots: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var b entity.Bot
		if err := rows.Scan(
			&b.Id,
			&b.Token,
			&b.Username,
			&b.Firstname,
			&b.IsDonor,
			&b.ChId,
			&b.ChLink,
			&b.GroupLinkId,
		); err != nil {
			return nil, fmt.Errorf("db: GetAllVampBots (2): %w", err)
		}
		bots = append(bots, b)
	}
	return bots, nil
}

func (s *Database) GetAllNoChannelBots() ([]entity.Bot, error) {
	bots := make([]entity.Bot, 0)
	q := `SELECT 
			id,
			token,
			username,
			first_name,
			is_donor,
			ch_id,
			ch_link,
			group_link_id
		FROM bots
		WHERE ch_id = 0`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("db: GetAllNoChannelBots: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var b entity.Bot
		if err := rows.Scan(
			&b.Id,
			&b.Token,
			&b.Username,
			&b.Firstname,
			&b.IsDonor,
			&b.ChId,
			&b.ChLink,
			&b.GroupLinkId,
		); err != nil {
			return nil, fmt.Errorf("db: GetAllNoChannelBots (2): %w", err)
		}
		bots = append(bots, b)
	}
	return bots, nil
}

func (s *Database) GetBotInfoById(botId int) (entity.Bot, error) {
	var b entity.Bot
	q := `SELECT
			id,
			token,
			username,
			first_name,
			is_donor,
			ch_id,
			ch_link,
			group_link_id
		FROM bots
		WHERE id = $1`
	err := s.db.QueryRow(q, botId).Scan(
		&b.Id,
		&b.Token,
		&b.Username,
		&b.Firstname,
		&b.IsDonor,
		&b.ChId,
		&b.ChLink,
		&b.GroupLinkId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return b, repository.ErrNotFound
			
		}
		return b, fmt.Errorf("db: GetBotInfoById: botId: %d err: %w", botId, err)
	}
	return b, nil
}

func (s *Database) GetBotInfoByToken(token string) (entity.Bot, error) {
	var b entity.Bot
	q := `SELECT
			id,
			token,
			username,
			first_name,
			is_donor,
			ch_id,
			ch_link,
			group_link_id
		FROM bots
		WHERE token = $1`
	err := s.db.QueryRow(q, token).Scan(
		&b.Id,
		&b.Token,
		&b.Username,
		&b.Firstname,
		&b.IsDonor,
		&b.ChId,
		&b.ChLink,
		&b.GroupLinkId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return b, fmt.Errorf("db: GetBotInfoByToken: token: %s ErrNotFound", token)
		}
		return b, fmt.Errorf("db: GetBotInfoByToken: token: %s err: %w", token, err)
	}
	return b, nil
}

func (s *Database) EditBotField(botId int, field string, content any) error {
	q := fmt.Sprintf(`UPDATE bots SET %s = $1 WHERE id = $2`, field)
	_, err := s.db.Exec(q, content, botId)
	if err != nil {
		return fmt.Errorf("db: EditBotField: botId: %d field: %s content: %v err: %w", botId, field, content, err)
	}
	return nil
}

func (s *Database) EditBotGroupLinkIdToNull(groupLinkId int) error {
	q := `UPDATE bots SET group_link_id = 0 WHERE group_link_id = $1`
	_, err := s.db.Exec(q, groupLinkId)
	if err != nil {
		return fmt.Errorf("db: EditBotGroupLinkIdToNull: groupLinkId: %d err: %w", groupLinkId, err)
	}
	return nil 
}

func (s *Database) EditBotGroupLinkId(groupLinkId, botId int) error {
	q := `UPDATE bots SET group_link_id = $1 WHERE id = $2`
	_, err := s.db.Exec(q, groupLinkId, botId)
	if err != nil {
		return fmt.Errorf("db: EditBotGroupLinkId: groupLinkId: %d botId: %d  err: %w", groupLinkId, botId, err)
		
	}
	return nil
}
