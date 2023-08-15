package tg_service

import (
	"fmt"
	"myapp/internal/models"
	u "myapp/internal/utils"
	"time"

	"go.uber.org/zap"
)

func (srv *TgService) HandleCallbackQuery(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.l.Info("tgClient: HandleCallbackQuery", zap.Any("cq", cq), zap.Any("chatId", chatId))

	if cq.Data == "create_vampere_bot" {
		err := srv.CQ_vampire_register(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "delete_vampere_bot" {
		err := srv.CQ_vampire_delete(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "add_ch_to_bot" {
		err := srv.CQ_add_ch_to_bot(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "create_group_link" {
		err := srv.CQ_create_group_link(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "update_group_link" {
		err := srv.CQ_update_group_link(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "delete_group_link" {
		err := srv.CQ_delete_group_link(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "add_admin_btn" {
		err := srv.CQ_add_admin(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "show_bots_and_channels" {
		err := srv.CQ_show_bots_and_channels(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "edit_bot_group_link" {
		err := srv.CQ_edit_bot_group_link(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "show_all_group_links" {
		err := srv.CQ_show_all_group_links(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "show_admin_panel" {
		err := srv.CQ_show_admin_panel(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "accept_ch_post_by_admin" {
		err := srv.CQ_accept_ch_post_by_admin(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "del_lost_bots" {
		err := srv.CQ_del_lost_bots(m)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
		}
		return err
	}

	if cq.Data == "restart_app" {
		srv.CQ_restart_app()
		return nil
	}

	return nil
}

func (srv *TgService) CQ_vampire_register(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, u.NEW_BOT_MSG)
	return err
}

func (srv *TgService) CQ_vampire_delete(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, u.DELETE_BOT_MSG)
	return err
}

func (srv *TgService) CQ_add_ch_to_bot(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, u.ADD_CH_TO_BOT_MSG)
	return err
}

func (srv *TgService) CQ_add_admin(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, u.NEW_ADMIN_MSG)
	return err
}

func (srv *TgService) CQ_show_bots_and_channels(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.showBotsAndChannels(chatId)
	return err
}

func (srv *TgService) CQ_edit_bot_group_link(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, u.EDIT_BOT_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_show_all_group_links(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.showAllGroupLinks(chatId)
	return err
}

func (srv *TgService) CQ_show_admin_panel(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.showAdminPanel(chatId)
	return err
}

func (srv *TgService) CQ_create_group_link(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, u.NEW_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_delete_group_link(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, u.DELETE_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_update_group_link(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, u.UPDATE_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_accept_ch_post_by_admin(m models.Update) error {
	// cq := m.CallbackQuery
	// chatId := cq.From.Id
	DonorBot, err := srv.As.GetBotInfoByToken(srv.Token)
	if err != nil {
		srv.l.Error("CQ_accept_ch_post_by_admin: srv.As.GetBotInfoByToken(srv.Token)", zap.Error(err))
	}
	srv.ShowMessClient(DonorBot.ChId, "ок, начинаю рассылку по остальным")
	srv.DeleteMess(DonorBot.ChId, m.CallbackQuery.Message.MessageId)

	go func(){
		err = srv.sendChPostAsVamp_Media_Group()
		if err != nil {
			srv.ShowMessClient(DonorBot.ChId, u.ERR_MSG +": " + err.Error())
		}
	}()

	return nil
}

func (srv *TgService) CQ_del_lost_bots(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	allBots, err := srv.As.GetAllBots()
	if err != nil {
		srv.l.Error("CQ_del_lost_bots: GetAllBots", zap.Error(err))
	}

	for _, bot := range allBots {
		if bot.IsDonor == 1 {
			continue
		}
		resp, err := srv.getBotByToken(bot.Token)
		if err != nil {
			srv.l.Error("CQ_del_lost_bots: getBotByToken", zap.Error(err), zap.Any("bot token", bot.Token))
		}
		if !resp.Ok && resp.ErrorCode == 401 && resp.Description == "Unauthorized" {
			err := srv.As.DeleteBot(bot.Id)
			if err != nil {
				srv.l.Error("CQ_del_lost_bots: DeleteBot", zap.Error(err), zap.Any("bot token", bot.Token))
			}
			srv.ShowMessClient(chatId, fmt.Sprintf("удален бот\nid: %d\nusername: %s\ntoken: %s\nканал id: %d\nканал link: %s", bot.Id, bot.Username, bot.Token, bot.ChId, bot.ChLink))
			time.Sleep(time.Second*3)
		}
	}

	srv.ShowMessClient(chatId, "проверка закончена")

	return nil
}

func (srv *TgService) CQ_restart_app() {
	go func(){
		time.Sleep(time.Second*3)
		panic("restart app")
	}()
}
