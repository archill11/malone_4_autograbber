package tg_service

import (
	"myapp/internal/models"

	"go.uber.org/zap"
)

func (srv *TgService) HandleNewChatMember(m models.Update) error {
	status := m.MyChatMember.NewChatMember.Status
	// chat := m.MyChatMember.Chat
	// newMemberId := m.MyChatMember.NewChatMember.User.Id
	// chatId := chat.Id
	// channelTitle := m.MyChatMember.Chat.Title

	if !m.MyChatMember.NewChatMember.User.IsBot {
		return nil
	}
	allBots, err := srv.As.GetAllBots()
	if err != nil {
		return err
	}
	ourBot := false
	for _, v := range allBots {
		if m.MyChatMember.NewChatMember.User.Id == v.Id {
			ourBot = true
			break
		}
	}
	if !ourBot {
		return nil
	}

	if status == "administrator" {
		err := srv.NCM_administrator(m)
		return err
	}

	return nil
}


func (srv *TgService) NCM_administrator(m models.Update) error {
	srv.l.Info("NCM_administrator:", zap.Any("models.Update", m))
	// status := m.MyChatMember.NewChatMember.Status
	chat := m.MyChatMember.Chat
	newMemberId := m.MyChatMember.NewChatMember.User.Id
	chatId := chat.Id
	// channelTitle := m.MyChatMember.Chat.Title

	bot, err := srv.As.GetBotInfoById(newMemberId)
	if err != nil {
		return err
	}
	cAny, err := srv.getChatByCurrBot(chatId, bot.Token)
	if err != nil {
		return err
	}
	bot.ChId = cAny.Result.Id
	bot.ChLink = cAny.Result.InviteLink

	err = srv.As.EditBotChField(bot)
	if err != nil {
		return err
	}

	return nil
}
