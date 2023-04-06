package tg

import "myapp/internal/models"

func (srv *TgClient) Donor_HandleChannelPost(m models.Update) error {
	// chatId := m.Message.Chat.Id
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	srv.l.Info("client::tg::HandleChannelPost::")

	err := srv.Ts.Donor_addChannelPost(m)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgClient) Vampire_HandleChannelPost(m models.Update) error {
	// chatId := m.Message.Chat.Id
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	srv.l.Info("client::tg::HandleChannelPost::")

	// err := srv.Ts.AddChannelPost(m)
	// if err != nil {
	// 	return err
	// }
	return nil
}
