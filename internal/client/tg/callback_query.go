package tg

import (
	"myapp/internal/models"
	u "myapp/internal/utils"

	"go.uber.org/zap"
)

func (srv *TgClient) HandleCallbackQuery(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tgClient: HandleCallbackQuery", zap.Any("cq", cq), zap.Any("chatId", chatId))

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

	if cq.Data == "create_group_link" {
		err := srv.Ts.CQ_create_group_link(m)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "update_group_link" {
		err := srv.Ts.CQ_update_group_link(m)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "delete_group_link" {
		err := srv.Ts.CQ_delete_group_link(m)
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

	if cq.Data == "show_all_group_links" {
		err := srv.Ts.CQ_show_all_group_links(m)
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

	if cq.Data == "accept_ch_post_by_admin" {
		err := srv.Ts.CQ_accept_ch_post_by_admin(m)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "restart_app" {
		srv.Ts.CQ_restart_app()
		return nil
	}

	return nil
}
