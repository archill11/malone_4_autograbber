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
		s.l.Err("Postgres: could not save the bot")
		return err
	} else {
		s.l.Info("Postgres: save bot")
	}
	return nil
}

func (s *Database) DeleteBot(id int) error {
	q := `DELETE FROM bots WHERE id = $1`
	_, err := s.db.Exec(q, id)
	if err != nil {
		s.l.Err("Postgres: ERR: could not delete the bot ")
		return err
	} else {
		s.l.Info("Postgres: delete bot")
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
			ch_link
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
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return b, repository.ErrNotFound
		}
		return b, err
	}
	return b, nil
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
			ch_link
		FROM bots
		WHERE is_donor = 0`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var b entity.Bot
		if err := rows.Scan(&b.Id, &b.Token, &b.Username, &b.Firstname, &b.IsDonor, &b.ChId, &b.ChLink); err != nil {
			return nil, err
		}
		bots = append(bots, b)
	}
	return bots, nil
}

func (s *Database) GetBotInfoById(botId int) (entity.Bot, error) {
	var b entity.Bot
	q := `SELECT id, username, first_name, token, is_donor
		FROM bots
		WHERE id = $1`
	err := s.db.QueryRow(q, botId).Scan(&b.Id, &b.Username, &b.Firstname, &b.Token, &b.IsDonor)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return b, repository.ErrNotFound
		}
		return b, err
	}
	return b, nil
}

func (s *Database) EditBotField(botId int, field string, content any) error {
	q := fmt.Sprintf(`UPDATE bots SET %s = $1 WHERE id = $2`, field)
	_, err := s.db.Exec(q, content, botId)
	if err != nil {
		s.l.Err("Postgres: could not change bot field ", field, content)
	} else {
		s.l.Info("Postgres: change bot field ", field, content)
	}
	return err
}
