package tg

import (
	"myapp/internal/models"
	u "myapp/internal/utils"
)

func (srv *TgClient) HandleCallbackQuery(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("client::tg::HandleCallbackQuery::", cq, chatId)

	if cq.Data == "create_vampere_bot" {
		err := srv.Ts.CQ_vampire_register(m)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "delete_vampere_bot" {
		err := srv.Ts.CQ_vampire_delete(m)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "add_admin_btn" {
		err := srv.Ts.CQ_add_admin(m)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "show_bots_and_channels" {
		err := srv.Ts.CQ_show_bots_and_channels(m)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "show_admin_panel" {
		err := srv.Ts.CQ_show_admin_panel(m)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	return nil
}
