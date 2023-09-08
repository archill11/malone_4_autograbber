package pg

import (
	"encoding/json"
	"fmt"
	"myapp/internal/entity"
)

func (s *Database) AddNewBot(id int, username, firstname, token string, isDonor int) error {
	q := `INSERT INTO bots (id, username, first_name, token, is_donor) 
			VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT DO NOTHING`
	_, err := s.Exec(q, id, username, firstname, token, isDonor)
	if err != nil {
		return fmt.Errorf("db: AddNewBot: %w", err)
	}
	return nil
}

func (s *Database) DeleteBot(id int) error {
	q := `DELETE FROM bots WHERE id = $1`
	_, err := s.Exec(q, id)
	if err != nil {
		return fmt.Errorf("db: DeleteBot: %w", err)
	}
	return nil
}

func (s *Database) GetBotByChannelId(channelId int) (entity.Bot, error) {
	q := `
		SELECT coalesce((
			SELECT to_json(c)
			FROM bots as c
			WHERE ch_id = $1 
		), '{}'::json)
	`
	var u entity.Bot
	var data []byte
	err := s.QueryRow(q, channelId).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetBotByChannelId Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetBotByChannelId Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetBotsByGrouLinkId(groupLinkId int) ([]entity.Bot, error) {
	q := `
		SELECT coalesce((
			SELECT json_agg(c)
			FROM bots as c
			WHERE group_link_id = $1 
		), '[]'::json)
	`
	u := make([]entity.Bot, 0)
	var data []byte
	err := s.QueryRow(q, groupLinkId).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetBotsByGrouLinkId Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetBotsByGrouLinkId Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetAllBots() ([]entity.Bot, error) {
	q := `
		SELECT coalesce((
			SELECT json_agg(c)
			FROM bots as c
		), '[]'::json)
	`
	u := make([]entity.Bot, 0)
	var data []byte
	err := s.QueryRow(q).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetAllBots Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetAllBots Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetAllVampBots() ([]entity.Bot, error) {
	q := `
		SELECT coalesce((
			SELECT json_agg(c)
			FROM bots as c
			WHERE is_donor = 0 
		), '[]'::json)
	`
	u := make([]entity.Bot, 0)
	var data []byte
	err := s.QueryRow(q).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetAllVampBots Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetAllVampBots Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetAllNoChannelBots() ([]entity.Bot, error) {
	q := `
		SELECT coalesce((
			SELECT json_agg(c)
			FROM bots as c
			WHERE ch_id = 0 
		), '[]'::json)
	`
	u := make([]entity.Bot, 0)
	var data []byte
	err := s.QueryRow(q).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetAllNoChannelBots Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetAllNoChannelBots Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetBotInfoById(botId int) (entity.Bot, error) {
	q := `
		SELECT coalesce((
			SELECT to_json(c)
			FROM bots as c
			WHERE id = $1 
		), '{}'::json)
	`
	var u entity.Bot
	var data []byte
	err := s.QueryRow(q, botId).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetBotInfoById Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetBotInfoById Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetBotInfoByToken(token string) (entity.Bot, error) {
	q := `
		SELECT coalesce((
			SELECT to_json(c)
			FROM bots as c
			WHERE token = $1 
		), '{}'::json)
	`
	var u entity.Bot
	var data []byte
	err := s.QueryRow(q, token).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetBotInfoByToken Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetBotInfoByToken Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) EditBotField(botId int, field string, content any) error {
	q := fmt.Sprintf(`UPDATE bots SET %s = $1 WHERE id = $2`, field)
	_, err := s.Exec(q, content, botId)
	if err != nil {
		return fmt.Errorf("db: EditBotField: botId: %d field: %s content: %v err: %w", botId, field, content, err)
	}
	return nil
}

func (s *Database) EditBotGroupLinkIdToNull(groupLinkId int) error {
	q := `UPDATE bots SET group_link_id = 0 WHERE group_link_id = $1`
	_, err := s.Exec(q, groupLinkId)
	if err != nil {
		return fmt.Errorf("db: EditBotGroupLinkIdToNull: groupLinkId: %d err: %w", groupLinkId, err)
	}
	return nil
}

func (s *Database) EditBotGroupLinkId(groupLinkId, botId int) error {
	q := `UPDATE bots SET group_link_id = $1 WHERE id = $2`
	_, err := s.Exec(q, groupLinkId, botId)
	if err != nil {
		return fmt.Errorf("db: EditBotGroupLinkId: groupLinkId: %d botId: %d  err: %w", groupLinkId, botId, err)

	}
	return nil
}
