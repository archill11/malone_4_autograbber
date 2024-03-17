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
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("HandleCallbackQuery: fromId: %d fromUsername: %s, cq.Data: %s", fromId, fromUsername, cq.Data))

	if cq.Data == "create_vampere_bot" {
		err := srv.CQ_vampire_register(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "delete_vampere_bot" {
		err := srv.CQ_vampire_delete(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "add_ch_to_bot" {
		err := srv.CQ_add_ch_to_bot(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "create_group_link" {
		err := srv.CQ_create_group_link(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "update_group_link" {
		err := srv.CQ_update_group_link(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "delete_group_link" {
		err := srv.CQ_delete_group_link(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "show_bots_and_channels" {
		err := srv.CQ_show_bots_and_channels(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "show_bots_and_channels_user" {
		err := srv.CQ_show_bots_and_channels_user(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "edit_bot_group_link" {
		err := srv.CQ_edit_bot_group_link(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "edit_bot_lichka" {
		err := srv.CQ_edit_bot_lichka(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "edit_bot_lichka_by_group_link" {
		err := srv.CQ_edit_bot_lichka_by_group_link(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "show_all_group_links" {
		err := srv.CQ_show_all_group_links(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "show_admin_panel" {
		err := srv.CQ_show_admin_panel(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "show_user_panel" {
		err := srv.CQ_show_user_panel(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "accept_ch_post_by_admin" {
		err := srv.CQ_accept_ch_post_by_admin(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "del_lost_bots" {
		err := srv.CQ_del_lost_bots(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "del_post_in_chs_bots" {
		err := srv.CQ_del_post_in_chs_bots(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "add_admin_btn" {
		err := srv.CQ_add_admin_btn(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "del_admin_btn" {
		err := srv.CQ_del_admin_btn(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "add_user_btn" {
		err := srv.CQ_add_user_btn(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if cq.Data == "del_user_btn" {
		err := srv.CQ_del_user_btn(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
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
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	return nil
}

func (srv *TgService) CQ_vampire_register(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_vampire_register: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.SendForceReply(fromId, NEW_BOT_MSG)
	return err
}

func (srv *TgService) CQ_vampire_delete(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_vampire_delete: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.SendForceReply(fromId, DELETE_BOT_MSG)
	return err
}

func (srv *TgService) CQ_add_ch_to_bot(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_add_ch_to_bot: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.SendForceReply(fromId, ADD_CH_TO_BOT_MSG)
	return err
}

func (srv *TgService) CQ_show_bots_and_channels(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_show_bots_and_channels: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.showBotsAndChannels(fromId)
	return err
}

func (srv *TgService) CQ_show_bots_and_channels_user(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_show_bots_and_channels_user: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.showBotsAndChannels_user(fromId)
	return err
}

func (srv *TgService) CQ_edit_bot_group_link(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_edit_bot_group_link: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.SendForceReply(fromId, EDIT_BOT_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_edit_bot_lichka(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_edit_bot_lichka: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.SendForceReply(fromId, EDIT_BOT_LICHKA_MSG)
	return err
}

func (srv *TgService) CQ_edit_bot_lichka_by_group_link(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_edit_bot_lichka_by_group_link: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.SendForceReply(fromId, EDIT_BOT_LICHKA_BY_GRLINK_MSG)
	return err
}

func (srv *TgService) CQ_show_all_group_links(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_show_all_group_links: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.showAllGroupLinks(fromId)
	return err
}

func (srv *TgService) CQ_show_admin_panel(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_show_admin_panel: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.showAdminPanel(fromId)
	return err
}

func (srv *TgService) CQ_show_user_panel(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_show_user_panel: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.showUserPanel(fromId)
	return err
}

func (srv *TgService) CQ_create_group_link(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_create_group_link: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.SendForceReply(fromId, NEW_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_delete_group_link(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_delete_group_link: fromId: %d fromUsername: %s", fromId, fromUsername))

	err := srv.SendForceReply(fromId, DELETE_GROUP_LINK_MSG)
	return err
}

func (srv *TgService) CQ_update_group_link(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_update_group_link: fromId: %d fromUsername: %s", fromId, fromUsername))

	srv.SendForceReply(fromId, UPDATE_GROUP_LINK_MSG)
	return nil
}

func (srv *TgService) CQ_accept_ch_post_by_admin(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_accept_ch_post_by_admin: fromId: %d fromUsername: %s", fromId, fromUsername))

	DonorBot, err := srv.db.GetBotInfoByToken(srv.Cfg.Token)
	if err != nil {
		return fmt.Errorf("CQ_accept_ch_post_by_admin GetBotInfoByToken token: %s err: %v", srv.Cfg.Token, err)
		
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
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_del_lost_bots: fromId: %d fromUsername: %s", fromId, fromUsername))

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
		resp, err := srv.GetMe(bot.Token)
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
			time.Sleep(time.Second)
		}
	}
	srv.SendMessage(fromId, "проверка закончена")
	return nil
}

func (srv *TgService) CQ_del_post_in_chs_bots(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_del_post_in_chs_bots: fromId: %d fromUsername: %s", fromId, fromUsername))

	srv.SendForceReply(fromId, DELETE_POST_MSG)
	return nil
}

func (srv *TgService) CQ_restart_app() {
	go func() {
		time.Sleep(time.Second)
		panic("restart app")
	}()
}

func (srv *TgService) CQ_edit_bot_group_link_stp2(m models.Update, botIdStr, grLinkIdStr string) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_edit_bot_group_link_stp2: fromId: %d fromUsername: %s", fromId, fromUsername))

	botId, err := strconv.Atoi(botIdStr)
	if err != nil {
		return fmt.Errorf("CQ_edit_bot_group_link_stp2: некоректный id бота: %s : %v", botIdStr, err)
	}
	groupLinkId, err := strconv.Atoi(grLinkIdStr)
	if err != nil {
		return fmt.Errorf("CQ_edit_bot_group_link_stp2: некоректный id ссылки: %s : %v", botIdStr, err)
	}
	err = srv.db.EditBotGroupLinkId(groupLinkId, botId)
	if err != nil {
		return fmt.Errorf("CQ_edit_bot_group_link_stp2: EditBotGroupLinkId err: %v", err)
	}
	srv.SendMessage(fromId, fmt.Sprintf("для бота %d, ссылка успешно изменена на %d", botId, groupLinkId))
	return nil
}

func (srv *TgService) CQ_add_admin_btn(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_add_admin_btn: fromId: %d fromUsername: %s", fromId, fromUsername))

	u, err := srv.db.GetUserById(fromId)
	if err != nil {
		return err
	}
	if u.Id == 0 {
		return nil
	}
	if u.IsAdmin == 0 {
		return nil
	}
	srv.SendForceReply(fromId, NEW_ADMIN_MSG)
	return nil
}

func (srv *TgService) CQ_del_admin_btn(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_del_admin_btn: fromId: %d fromUsername: %s", fromId, fromUsername))

	u, err := srv.db.GetUserById(fromId)
	if err != nil {
		return err
	}
	if u.Id == 0 {
		return nil
	}
	if u.IsAdmin == 0 {
		return nil
	}

	srv.SendForceReply(fromId, DEL_ADMIN_MSG)
	return nil
}

func (srv *TgService) CQ_add_user_btn(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_add_user_btn: fromId: %d fromUsername: %s", fromId, fromUsername))

	u, err := srv.db.GetUserById(fromId)
	if err != nil {
		return err
	}
	if u.Id == 0 {
		return nil
	}
	if u.IsAdmin == 0 {
		return nil
	}
	srv.SendForceReply(fromId, NEW_USER_MSG)
	return nil
}

func (srv *TgService) CQ_del_user_btn(m models.Update) error {
	cq := m.CallbackQuery
	fromId := cq.From.Id
	fromUsername := cq.From.UserName
	srv.l.Info(fmt.Sprintf("CQ_del_user_btn: fromId: %d fromUsername: %s", fromId, fromUsername))

	u, err := srv.db.GetUserById(fromId)
	if err != nil {
		return err
	}
	if u.Id == 0 {
		return nil
	}
	if u.IsAdmin == 0 {
		return nil
	}

	srv.SendForceReply(fromId, DEL_USER_MSG)
	return nil
}