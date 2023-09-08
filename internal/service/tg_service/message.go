package tg_service

import (
	"fmt"
	"myapp/internal/models"

	"go.uber.org/zap"
)

func (srv *TgService) HandleMessage(m models.Update) error {
	// chatId := m.Message.Chat.Id
	// userFirstName := m.Message.From.FirstName
	userUserName := m.Message.From.UserName
	msgText := m.Message.Text

	srv.l.Info("tgClient: HandleMessage", zap.Any("userUserName", userUserName), zap.Any("msgText", msgText))

	if msgText == "/admin" {
		err := srv.M_admin(m)
		return err
	}

	if msgText == "/start" {
		err := srv.M_start(m)
		return err
	}

	return nil
}

func (srv *TgService) M_start(m models.Update) error {
	chatId := m.Message.Chat.Id
	msgText := m.Message.Text
	userFirstName := m.Message.From.FirstName
	userUserName := m.Message.From.UserName
	srv.l.Info("M_start:", zap.Any("userUserName", userUserName), zap.Any("msgText", msgText))

	err := srv.SendMessage(chatId, fmt.Sprintf("Привет %s", userFirstName))
	if err != nil {
		return err
	}
	err = srv.db.AddNewUser(chatId, userUserName, userFirstName)

	return err
}

func (srv *TgService) M_admin(m models.Update) error {
	chatId := m.Message.Chat.Id
	msgText := m.Message.Text
	userUserName := m.Message.From.UserName
	srv.l.Info("M_admin:", zap.Any("userUserName", userUserName), zap.Any("msgText", msgText))

	u, err := srv.db.GetUserById(chatId)
	if err != nil {
		return err
	}
	if u.Id == 0 {
		srv.SendMessage(chatId, "Нажмите сначала /start")
		return nil
	}
	if u.IsAdmin == 0 {
		return nil
	}
	err = srv.showAdminPanel(chatId)

	return err
}
