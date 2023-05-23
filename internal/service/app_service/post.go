package service

import "myapp/internal/entity"

func (s *AppService) AddNewPost(channelId, postId, donorChPostId int) error {
	u := entity.NewPost(channelId, postId, donorChPostId)
	return s.db.AddNewPost(u)
}

func (s *AppService) GetPostByDonorIdAndChId(donorChPostId, channelId int) (entity.Post, error) {
	return s.db.GetPostByDonorIdAndChId(donorChPostId, channelId)
}

func (s *AppService) GetPostByChIdAndBotToken(channelId int, botToken string) (entity.Post, error) {
	return s.db.GetPostByChIdAndBotToken(channelId, botToken)
}

// func (s *AppService) GetChIdByBotToken(botToken string) (entity.Post, error) {
// 	uf, err := s.db.GetChIdByBotToken(botToken)
// 	if err != nil {
// 		return uf, err
// 	}
// 	return uf, nil
// }
