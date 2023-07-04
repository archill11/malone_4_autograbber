package pg

import (
	"database/sql"
	"errors"
	"fmt"
	"myapp/internal/entity"
	"myapp/internal/repository"
)

func (s *Database) AddNewPost(u entity.Post) error {
	q := `INSERT INTO posts 
		(ch_id, post_id, donor_ch_post_id) 
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`
	_, err := s.db.Exec(q, u.ChId, u.PostId, u.DonorChPostId)
	if err != nil {
		return fmt.Errorf("db: AddNewPost: ChId: %d PostId %d DonorChPostId %d err: %w", u.ChId, u.PostId, u.DonorChPostId, err)
	}
	return nil
}

func (s *Database) GetPostByDonorIdAndChId(donorChPostId, channelId int) (entity.Post, error) {
	var p entity.Post
	q := `
		SELECT
			ch_id,
			post_id,
			donor_ch_post_id
		FROM posts
		WHERE ch_id = $1 
		AND donor_ch_post_id = $2`
	err := s.db.QueryRow(q, channelId, donorChPostId).Scan(
		&p.ChId,
		&p.PostId,
		&p.DonorChPostId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p, repository.ErrNotFound
		}
		return p, fmt.Errorf("db: GetPostByDonorIdAndChId: channelId: %d donorChPostId %d err: %w", channelId, donorChPostId, err)
	}
	return p, nil
}

func (s *Database) GetPostByChIdAndBotToken(channelId int, botToken string) (entity.Post, error) {
	var p entity.Post
	q := `
		SELECT
			p.ch_id,
			p.post_id,
			p.donor_ch_post_id,
		FROM posts AS p
		JOIN bots AS b
			ON p.ch_id = b.ch_id
		WHERE p.ch_id = $1 
		AND b.token = $2`
	err := s.db.QueryRow(q, channelId, botToken).Scan(
		&p.ChId,
		&p.PostId,
		&p.DonorChPostId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p, fmt.Errorf("db: GetPostByChIdAndBotToken: channelId: %d botToken %s ErrNoRows", channelId, botToken)
		}
		return p, fmt.Errorf("db: GetPostByChIdAndBotToken: channelId: %d botToken %s err: %w", channelId, botToken, err)
	}
	return p, nil
}

// func (s *Database) GetChByBotToken(botToken string) (entity.Post, error) {
// 	var p entity.Post
// 	q := `
// 		SELECT
// 			ch_id,
// 			post_id,
// 			donor_ch_post_id,
// 			bot_token
// 		FROM posts
// 		WHERE bot_token = $1`
// 	err := s.db.QueryRow(q, botToken).Scan(
// 		&p.ChId,
// 		&p.PostId,
// 		&p.DonorChPostId,
// 		&p.BotToken,
// 	)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return p, repository.ErrNotFound
// 		}
// 		return p, err
// 	}
// 	return p, nil
// }
