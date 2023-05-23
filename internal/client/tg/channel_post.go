package tg

import (
	"myapp/internal/models"
	u "myapp/internal/utils"
)

func (srv *TgClient) Donor_HandleChannelPost(m models.Update) error {
	chatId := m.ChannelPost.Chat.Id
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	srv.l.Info("client::tg::HandleChannelPost::")
	srv.l.Warn("new event ")

	err := srv.Ts.Donor_addChannelPost(m)
	if err != nil {
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG_2+err.Error())
		}
		return err
	}
	return nil
}

// func (srv *TgClient) Vampire_HandleChannelPost(m models.Update) error {
// 	// chatId := m.Message.Chat.Id
// 	// msgText := m.Message.Text
// 	// userFirstName := m.Message.From.FirstName
// 	// userUserName := m.Message.From.UserName
// 	srv.l.Info("client::tg::HandleChannelPost::")

// 	// err := srv.Ts.AddChannelPost(m)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	return nil
// }
