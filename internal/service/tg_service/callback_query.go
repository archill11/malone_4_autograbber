package tg_service

import (
	"bytes"
	"fmt"
	"myapp/internal/models"
	my_regex "myapp/pkg/regex"
	"strconv"
	"strings"
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
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "delete_vampere_bot" {
		err := srv.CQ_vampire_delete(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "add_ch_to_bot" {
		err := srv.CQ_add_ch_to_bot(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "create_group_link" {
		err := srv.CQ_create_group_link(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "update_group_link" {
		err := srv.CQ_update_group_link(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "delete_group_link" {
		err := srv.CQ_delete_group_link(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "add_admin_btn" {
		err := srv.CQ_add_admin(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "show_bots_and_channels" {
		err := srv.CQ_show_bots_and_channels(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "edit_bot_group_link" {
		err := srv.CQ_edit_bot_group_link(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "show_all_group_links" {
		err := srv.CQ_show_all_group_links(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "show_admin_panel" {
		err := srv.CQ_show_admin_panel(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "accept_ch_post_by_admin" {
		err := srv.CQ_accept_ch_post_by_admin(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "del_lost_bots" {
		err := srv.CQ_del_lost_bots(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "del_post_in_chs_bots" {
		err := srv.CQ_del_post_in_chs_bots(m)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	if cq.Data == "restart_app" {
		srv.CQ_restart_app()
		return nil
	}

	if strings.HasPrefix(cq.Data, "edit_bot_") { // edit_bot_%s_link_to_%d_gr_link_btn
		botId := my_regex.GetStringInBetween(cq.Data, "edit_bot_", "_link")
		grLinkId := my_regex.GetStringInBetween(cq.Data, "to_", "_gr_link")
		err := srv.CQ_edit_bot_group_link_stp2(m, botId, grLinkId)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}

	return nil
}

func (srv *TgService) CQ_vampire_register(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, NEW_BOT_MSG)
	return err
}

func (srv *TgService) CQ_vampire_delete(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, DELETE_BOT_MSG)
	return err
}

func (srv *TgService) CQ_add_ch_to_bot(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, ADD_CH_TO_BOT_MSG)
	return err
}

func (srv *TgService) CQ_add_admin(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, NEW_ADMIN_MSG)
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
	err := srv.SendForceReply(chatId, EDIT_BOT_GROUP_LINK_MSG)
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
	err := srv.SendForceReply(chatId, NEW_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_delete_group_link(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	err := srv.SendForceReply(chatId, DELETE_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_update_group_link(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_update_group_link: fromId-%d fromUsername-%s", fromId, fromUsername))

	srv.SendForceReply(fromId, UPDATE_GROUP_LINK_MSG)
	return nil
}

func (srv *TgService) CQ_accept_ch_post_by_admin(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_accept_ch_post_by_admin: fromId-%d fromUsername-%s", fromId, fromUsername))

	DonorBot, err := srv.db.GetBotInfoByToken(srv.Cfg.Token)
	if err != nil {
		return fmt.Errorf("CQ_accept_ch_post_by_admin GetBotInfoByToken token-%s err: %v", srv.Cfg.Token, err)
		
	}
	srv.SendMessage(DonorBot.ChId, "ок, начинаю рассылку по остальным")
	srv.DeleteMessage(DonorBot.ChId, m.CallbackQuery.Message.MessageId, srv.Cfg.Token)

	go func() {
		err = srv.sendChPostAsVamp_Media_Group()
		if err != nil {
			srv.SendMessage(DonorBot.ChId, ERR_MSG)
			srv.SendMessage(DonorBot.ChId, err.Error())
		}
	}()

	return nil
}

func (srv *TgService) CQ_del_lost_bots(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	allBots, err := srv.db.GetAllBots()
	if err != nil {
		errMess := fmt.Sprintf("CQ_del_lost_bots: GetAllBots err: %v", err)
		srv.l.Error(errMess)
	}
	if len(allBots) == 0 {
		errMess := fmt.Sprintf("CQ_del_lost_bots: GetAllBots err: len(allBots) == 0")
		srv.l.Error(errMess)
	}

	for _, bot := range allBots {
		if bot.IsDonor == 1 {
			continue
		}
		resp, err := srv.getBotByToken(bot.Token)
		if err != nil {
			errMess := fmt.Sprintf("CQ_del_lost_bots: getBotByToken err: %v", err)
			srv.l.Error(errMess, zap.Any("bot token", bot.Token))
		}
		if !resp.Ok && resp.ErrorCode == 401 && resp.Description == "Unauthorized" {
			srv.db.DeleteBot(bot.Id)

			var mess bytes.Buffer
			mess.WriteString("удален бот без доступа\n")
			mess.WriteString(fmt.Sprintf("бот: @%s | %s\n", bot.Username, bot.Token))
			mess.WriteString(fmt.Sprintf("канал: %d | %s\n", bot.ChId, bot.ChLink))
			logMess := mess.String()

			srv.SendMessage(fromId, logMess)
			time.Sleep(time.Second * 3)
		}
	}
	srv.SendMessage(fromId, "проверка закончена")
	return nil
}

func (srv *TgService) CQ_del_post_in_chs_bots(m models.Update) error {
	cq := m.CallbackQuery
	chatId := cq.From.Id
	srv.SendForceReply(chatId, DELETE_POST_MSG)
	return nil
}

func (srv *TgService) CQ_restart_app() {
	go func() {
		time.Sleep(time.Second * 3)
		panic("restart app")
	}()
}

func (srv *TgService) CQ_edit_bot_group_link_stp2(m models.Update, botIdStr, grLinkIdStr string) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	srv.l.Info("tg_service: CQ_edit_bot_group_link_stp2", zap.Any("fromId", fromId))

	botId, err := strconv.Atoi(botIdStr)
	if err != nil {
		return fmt.Errorf("CQ_edit_bot_group_link_stp2: некоректный id бота-%s : %v", botIdStr, err)
	}
	groupLinkId, err := strconv.Atoi(grLinkIdStr)
	if err != nil {
		return fmt.Errorf("CQ_edit_bot_group_link_stp2: некоректный id ссылки-%s : %v", botIdStr, err)
	}
	err = srv.db.EditBotGroupLinkId(groupLinkId, botId)
	if err != nil {
		return fmt.Errorf("CQ_edit_bot_group_link_stp2: EditBotGroupLinkId err: %v", err)
	}
	srv.SendMessage(fromId, fmt.Sprintf("для бота %d, ссылка успешно изменена на %d", botId, groupLinkId))
	return nil
}