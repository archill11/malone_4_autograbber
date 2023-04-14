package service

import (
	"myapp/internal/entity"
)

func (s *AppService) AddNewBot(id int, username, firstname, token string, idDonor int) error {
	err := s.db.AddNewBot(id, username, firstname, token, idDonor)
	return err
}

func (s *AppService) DeleteBot(botId int) error {
	err := s.db.DeleteBot(botId)
	return err
}

func (s *AppService) GetAllBots() ([]entity.Bot, error) {
	bots, err := s.db.GetAllBots()
	return bots, err
}

func (s *AppService) GetAllVampBots() ([]entity.Bot, error) {
	bots, err := s.db.GetAllVampBots()
	return bots, err
}

func (s *AppService) GetBotByChannelId(chatId int) (entity.Bot, error) {
	bot, err := s.db.GetBotByChannelId(chatId)
	return bot, err
}

func (s *AppService) GetBotInfoById(botId int) (entity.Bot, error) {
	bot, err := s.db.GetBotInfoById(botId)
	return bot, err
}

func (s *AppService) EditBotChField(bot entity.Bot) error {
	err := s.db.EditBotField(bot.Id, "ch_id", bot.ChId)
	if err != nil {
		return err
	}
	err = s.db.EditBotField(bot.Id, "ch_link", bot.ChLink)
	if err != nil {
		return err
	}
	return nil
}

func (s *AppService) EditBotGroupLinkId(groupLinkId int) error {
	err := s.db.EditBotGroupLinkId(groupLinkId)
	return err
}
