package pg

import (
	"encoding/json"
	"fmt"
	"myapp/internal/entity"
)

func (s *Database) AddNewPost(chId, postId, donorChPostId int, caption string) error {
	q := `INSERT INTO posts (ch_id, post_id, donor_ch_post_id, caption) 
			VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`
	_, err := s.Exec(q, chId, postId, donorChPostId, caption)
	if err != nil {
		return fmt.Errorf("db: AddNewPost: ChId: %d PostId %d DonorChPostId %d err: %w", chId, postId, donorChPostId, err)
	}
	return nil
}

func (s *Database) GetPostByDonorIdAndChId(donorChPostId, channelId int) (entity.Post, error) {
	q := `
		SELECT coalesce((
			SELECT to_json(c)
			FROM posts as c
			WHERE ch_id = $1 
			AND donor_ch_post_id = $2
		), '{}'::json)
	`
	var u entity.Post
	var data []byte
	err := s.QueryRow(q, channelId, donorChPostId).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetPostByDonorIdAndChId Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetPostByDonorIdAndChId Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetPostsByDonorIdAndChId(donorChPostId, channelId int) ([]entity.Post, error) {
	q := `
		SELECT coalesce((
			SELECT json_agg(c)
			FROM posts as c
			WHERE ch_id = $1 
			AND donor_ch_post_id = $2
		), '[]'::json)
	`
	u := make([]entity.Post, 0)
	var data []byte
	err := s.QueryRow(q, channelId, donorChPostId).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetPostsByDonorIdAndChId Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetPostsByDonorIdAndChId Unmarshal: %v", err)
	}
	return u, nil
}

func (s *Database) GetPostByChIdAndBotToken(channelId int, botToken string) (entity.Post, error) {
	q := `
		SELECT coalesce((
			SELECT to_json(c)
			FROM posts as p
			JOIN bots AS b
				ON p.ch_id = b.ch_id
			WHERE p.ch_id = $1 
			AND b.token = $2
		), '{}'::json)
	`
	var u entity.Post
	var data []byte
	err := s.QueryRow(q, channelId, botToken).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetPostByDonorIdAndChId Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetPostByDonorIdAndChId Unmarshal: %v", err)
	}
	return u, nil
}
