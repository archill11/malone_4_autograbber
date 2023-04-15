package tg_service

import (
	"fmt"
	"myapp/internal/models"
)

func (srv *TgService) NCM_administrator(m models.Update) error {
	srv.l.Info("tg_service::tg::newChatMem::administrator::", m)
	fmt.Println("tg_service::tg::newChatMem::administrator::", m)
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
	srv.l.Info("cAny::Result:", cAny)
	bot.ChId = cAny.Result.Id
	bot.ChLink = cAny.Result.InviteLink

	err = srv.As.EditBotChField(bot)
	if err != nil {
		return err
	}

	return nil
}
