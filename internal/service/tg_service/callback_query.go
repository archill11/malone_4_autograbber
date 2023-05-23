package tg_service

import (
	"myapp/internal/models"
	u "myapp/internal/utils"
)

func (srv *TgService) CQ_vampire_register(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.SendForceReply(chatId, u.NEW_BOT_MSG)
	return err
}

func (srv *TgService) CQ_vampire_delete(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.SendForceReply(chatId, u.DELETE_BOT_MSG)
	return err
}

func (srv *TgService) CQ_add_admin(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.SendForceReply(chatId, u.NEW_ADMIN_MSG)
	return err
}

func (srv *TgService) CQ_show_bots_and_channels(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.showBotsAndChannels(chatId)
	return err
}

func (srv *TgService) CQ_show_all_group_links(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.showAllGroupLinks(chatId)
	return err
}

func (srv *TgService) CQ_show_admin_panel(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.showAdminPanel(chatId)
	return err
}

func (srv *TgService) CQ_create_group_link(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.SendForceReply(chatId, u.NEW_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_delete_group_link(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.SendForceReply(chatId, u.DELETE_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_update_group_link(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.SendForceReply(chatId, u.UPDATE_GROUP_LINK_MSG)
	return err
}
