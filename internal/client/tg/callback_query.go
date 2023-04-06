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
		err := srv.Ts.CQ_vampire_register(m)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	return nil
}
