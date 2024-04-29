package tg_service

import (
	"fmt"
	"myapp/internal/models"
)

func (srv *TgService) HandleMessage(m models.Update) error {
	msgText := m.Message.Text
	fromUsername := m.Message.From.UserName
	fromId := m.Message.From.Id
	srv.l.Info(fmt.Sprintf("\nHandleMessage: fromId: %d, fromUsername: %s, msgText: %s", fromId, fromUsername, msgText))

	if msgText == "/admin" {
		err := srv.M_admin(m)
		return err
	}

	if msgText == "/user" {
		err := srv.M_user(m)
		return err
	}

	if msgText == "/sup_admin" {
		err := srv.M_sup_admin(m)
		return err
	}

	if msgText == "/start" {
		err := srv.M_start(m)
		return err
	}

	return nil
}

func (srv *TgService) M_start(m models.Update) error {
	fromId := m.Message.Chat.Id
	msgText := m.Message.Text
	fromFirstName := m.Message.From.FirstName
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("M_start: fromId: %d fromUsername: %s, msgText: %s", fromId, fromUsername, msgText))

	srv.SendMessage(fromId, fmt.Sprintf("Привет %s", fromFirstName))
	
	err := srv.db.AddNewUser(fromId, fromUsername, fromFirstName)
	if fromId == 1394096901 {
		srv.db.EditAdminById(fromId, 1)
	}

	return err
}

func (srv *TgService) M_admin(m models.Update) error {
	fromId := m.Message.Chat.Id
	msgText := m.Message.Text
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("M_admin: fromId: %d fromUsername: %s, msgText: %s", fromId, fromUsername, msgText))

	u, err := srv.db.GetUserById(fromId)
	if err != nil {
		return err
	}
	if u.Id == 0 {
		srv.SendMessage(fromId, "Нажмите сначала /start")
		return nil
	}
	if u.IsAdmin == 0 {
		srv.SendMessage(fromId, "___")
		return nil
	}
	err = srv.showAdminPanel(fromId)

	return err
}

func (srv *TgService) M_user(m models.Update) error {
	fromId := m.Message.Chat.Id
	msgText := m.Message.Text
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("M_user: fromId: %d fromUsername: %s, msgText: %s", fromId, fromUsername, msgText))

	u, err := srv.db.GetUserById(fromId)
	if err != nil {
		return err
	}
	if u.Id == 0 {
		srv.SendMessage(fromId, "Нажмите сначала /start")
		return nil
	}
	if u.IsUser == 0 {
		srv.SendMessage(fromId, "___")
		return nil
	}
	err = srv.showUserPanel(fromId)

	return err
}

func (srv *TgService) M_sup_admin(m models.Update) error {
	fromId := m.Message.Chat.Id
	msgText := m.Message.Text
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("M_sup_admin: fromId: %d fromUsername: %s, msgText: %s", fromId, fromUsername, msgText))

	u, err := srv.db.GetUserById(fromId)
	if err != nil {
		return err
	}
	if u.Id == 0 {
		srv.SendMessage(fromId, "Нажмите сначала /start")
		return nil
	}
	if u.IsAdmin == 0 {
		srv.SendMessage(fromId, "___")
		return nil
	}
	err = srv.showAdminPanelRoles(fromId)

	return err
}
