package tg_service

import (
	"myapp/internal/models"
	u "myapp/internal/utils"
)

func (srv *TgService) CQ_vampire_register(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.sendForceReply(chatId, u.NEW_BOT_MSG)
	return err
}

func (srv *TgService) CQ_vampire_delete(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tg_service::tg::cq::", cq.Data, chatId)

	err := srv.sendForceReply(chatId, u.DELETE_BOT_MSG)
	return err
}
