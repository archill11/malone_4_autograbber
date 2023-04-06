package tg

import (
	"fmt"
	"myapp/internal/models"
)

func (srv *TgClient) HandleNewChatMember(m models.Update) error {
	srv.l.Info("client::tg::newChatMem::", m)
	fmt.Println("client::tg::newChatMem::", m)
	status := m.MyChatMember.NewChatMember.Status
	// chat := m.MyChatMember.Chat
	// newMemberId := m.MyChatMember.NewChatMember.User.ID
	// chatId := chat.ID
	// channelTitle := m.MyChatMember.Chat.Title

	if !m.MyChatMember.NewChatMember.User.IsBot {
		return nil
	}

	if status == "administrator" {
		err := srv.Ts.NCM_administrator(m)
		return err
	}

	return nil
}
