package tg

import (
	"myapp/internal/models"

	"go.uber.org/zap"
)

func (srv *TgClient) HandleMessage(m models.Update) error {
	// chatId := m.Message.Chat.Id
	// userFirstName := m.Message.From.FirstName
	userUserName := m.Message.From.UserName
	msgText := m.Message.Text

	srv.l.Info("tgClient: HandleMessage", zap.Any("userUserName", userUserName), zap.Any("msgText", msgText))

	if msgText == "/admin" {
		err := srv.Ts.M_admin(m)
		return err
	}

	if msgText == "/start" {
		err := srv.Ts.M_start(m)
		return err
	}

	return nil
}
