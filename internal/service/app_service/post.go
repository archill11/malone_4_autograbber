package service

import "myapp/internal/entity"

func (s *AppService) AddNewPost(channelId, postId, donorChPostId int) error {
	u := entity.NewPost(channelId, postId, donorChPostId)
	err := s.db.AddNewPost(u)
	if err != nil {
		return err
	}
	return nil
}

func (s *AppService) GetPostByDonorIdAndChId(donorChPostId, channelId int) (entity.Post, error) {
	uf, err := s.db.GetPostByDonorIdAndChId(donorChPostId, channelId)
	if err != nil {
		return uf, err
	}
	return uf, nil
}

func (s *AppService) GetPostByChIdAndBotToken(channelId int, botToken string) (entity.Post, error) {
	uf, err := s.db.GetPostByChIdAndBotToken(channelId, botToken)
	if err != nil {
		return uf, err
	}
	return uf, nil
}

// func (s *AppService) GetChIdByBotToken(botToken string) (entity.Post, error) {
// 	uf, err := s.db.GetChIdByBotToken(botToken)
// 	if err != nil {
// 		return uf, err
// 	}
// 	return uf, nil
// }
