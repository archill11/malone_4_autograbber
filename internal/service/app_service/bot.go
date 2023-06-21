package service

import (
	"myapp/internal/entity"
)

func (s *AppService) AddNewBot(id int, username, firstname, token string, idDonor int) error {
	return s.db.AddNewBot(id, username, firstname, token, idDonor)
}

func (s *AppService) DeleteBot(botId int) error {
	return s.db.DeleteBot(botId)
}

func (s *AppService) GetAllBots() ([]entity.Bot, error) {
	return s.db.GetAllBots()
}

func (s *AppService) GetAllVampBots() ([]entity.Bot, error) {
	return s.db.GetAllVampBots()
}

func (s *AppService) GetAllNoChannelBots() ([]entity.Bot, error) {
	return s.db.GetAllNoChannelBots()
}

func (s *AppService) GetBotByChannelId(chatId int) (entity.Bot, error) {
	return s.db.GetBotByChannelId(chatId)
}

func (s *AppService) GetBotsByGrouLinkId(groupLinkId int) ([]entity.Bot, error) {
	return s.db.GetBotsByGrouLinkId(groupLinkId)
}

func (s *AppService) GetBotInfoById(botId int) (entity.Bot, error) {
	return s.db.GetBotInfoById(botId)
}

func (s *AppService) GetBotInfoByToken(token string) (entity.Bot, error) {
	return s.db.GetBotInfoByToken(token)
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

func (s *AppService) EditBotGroupLinkIdToNull(groupLinkId int) error {
	return s.db.EditBotGroupLinkIdToNull(groupLinkId)
}

func (s *AppService) EditBotGroupLinkId(groupLinkId, botId int) error {
	return s.db.EditBotGroupLinkId(groupLinkId, botId)
}
