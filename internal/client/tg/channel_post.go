package tg

import (
	"myapp/internal/models"
	u "myapp/internal/utils"

	"go.uber.org/zap"
)

func (srv *TgClient) Donor_HandleChannelPost(m models.Update) error {
	chatId := m.ChannelPost.Chat.Id
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	srv.l.Info("tgClient: Donor_HandleChannelPost", zap.Any("models.Update", m))

	err := srv.Ts.Donor_addChannelPost(m)
	if err != nil {
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG_2 + err.Error())
		}
		return err
	}
	return nil
}

