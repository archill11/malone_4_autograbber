package tg_service

import (
	"fmt"
	"myapp/internal/models"
)

func (srv *TgService) Donor_addChannelPost(m models.Update) error {
	// chatId := m.Message.Chat.ID
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	srv.l.Info("tg_service::AddChannelPost::")

	message_id := m.ChannelPost.MessageId
	channel_id := m.ChannelPost.Chat.Id
	fmt.Println("|_|")

	// добавили пост в БД
	err := srv.As.AddNewPost(channel_id, message_id, message_id)
	if err != nil {
		return err
	}

	allVampBots, err := srv.As.GetAllVampBots()
	if err != nil {
		return err
	}

	for _, vampBot := range allVampBots {
		srv.sendChPostAsVamp(vampBot, m)
	}

	return nil
}
