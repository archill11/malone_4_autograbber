package tg_service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"myapp/internal/entity"
	"myapp/internal/models"
	"myapp/internal/repository"
	u "myapp/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

func (srv *TgService) HandleReplyToMessage(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	// chatId := m.Message.From.Id

	srv.l.Info("tgClient: HandleReplyToMessage", zap.Any("rm.Tex", rm.Text), zap.Any("replyMes", replyMes))

	if rm.Text == u.NEW_BOT_MSG {
		err := srv.RM_obtain_vampire_bot_token(m)
		return err
	}

	if rm.Text == u.DELETE_BOT_MSG {
		err := srv.RM_delete_bot(m)
		return err
	}

	if rm.Text == u.ADD_CH_TO_BOT_MSG {
		err := srv.RM_add_ch_to_bot(m)
		return err
	}

	if strings.HasPrefix(rm.Text, "укажите id канала в котором уже бот админ и к которому нужно привязать бота-") {
		runes := []rune(rm.Text)
		runesStr := string(runes[len([]rune("укажите id канала в котором уже бот админ и к которому нужно привязать бота-")):])
		botId, _ := strconv.Atoi(runesStr)
		err := srv.RM_add_ch_to_bot_spet2(m, botId)
		return err
	}

	if rm.Text == u.NEW_ADMIN_MSG {
		err := srv.RM_add_admin(m)
		return err
	}

	if rm.Text == u.NEW_GROUP_LINK_MSG {
		err := srv.RM_add_group_link(m)
		return err
	}

	if rm.Text == u.EDIT_BOT_GROUP_LINK_MSG {
		err := srv.RM_edit_bot_group_link(m)
		return err
	}

	if rm.Text == u.DELETE_GROUP_LINK_MSG {
		err := srv.RM_delete_group_link(m)
		return err
	}

	if rm.Text == u.UPDATE_GROUP_LINK_MSG {
		chatId := m.Message.From.Id
		replyMes := m.Message.Text
		replyMes = strings.TrimSpace(replyMes)
	
		grId, err := strconv.Atoi(replyMes)
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG)
			return err
		}
		err = srv.SendForceReply(chatId, fmt.Sprintf(u.GROUP_LINK_UPDATE_MSG, grId))
		return err
	}

	if strings.HasPrefix(rm.Text, "укажите номер группы-ссылки для нового бота[") {
		runes := []rune(rm.Text)
		runesStr := string(runes[len([]rune("укажите номер группы-ссылки для нового бота[")):])
		botId, _ := strconv.Atoi(runesStr)
		err := srv.RM_update_bot_group_link(m, botId)
		return err
	}

	if strings.HasPrefix(rm.Text, "укажите новую ссылку для ref [") {
		runes := []rune(rm.Text)
		runesStr := string(runes[len([]rune("укажите новую ссылку для ref [")):])
		refId, _ := strconv.Atoi(runesStr)
		err := srv.RM_update_group_link(m, refId)
		return err
	}

	return nil
}

func (srv *TgService) RM_obtain_vampire_bot_token(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_obtain_vampire_bot_token", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	tgobotResp, err := srv.getBotByToken(strings.TrimSpace(replyMes))
	if err != nil {
		return err
	}
	res := tgobotResp.Result
	bot := entity.NewBot(res.Id, res.UserName, res.FirstName, strings.TrimSpace(replyMes), 0)
	err = srv.As.AddNewBot(bot.Id, bot.Username, bot.Firstname, bot.Token, bot.IsDonor)
	if err != nil {
		return err
	}
	// tgResp := struct {
	// 	Ok          bool   `json:"ok"`
	// 	Description string `json:"description"`
	// }{}
	// resp, err := http.Get(fmt.Sprintf(
	// 	srv.TgEndp, bot.Token, fmt.Sprintf("setWebhook?url=%s/api/v1/vampire/update", srv.HostUrl)), // set Webhook
	// )
	// if err != nil {
	// 	srv.l.Error("RM_obtain_vampire_bot_token: set Webhook::", zap.Error(err))
	// }
	// defer resp.Body.Close()

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// err = json.Unmarshal(body, &tgResp)
	// if err != nil {
	// 	srv.ShowMessClient(chatId, u.ERR_MSG)
	// 	return err
	// }
	// if !tgResp.Ok {
	// 	srv.l.Error("RM_obtain_vampire_bot_token: !tgResp.Ok", zap.Error(err), zap.Any("tgResp.Description", tgResp.Description))
	// 	srv.ShowMessClient(chatId, u.ERR_MSG)
	// }
	// srv.l.Info("RM_obtain_vampire_bot_token: set Webhook", zap.Any("Webhook url", fmt.Sprintf(srv.TgEndp, bot.Token, fmt.Sprintf("setWebhook?url=%s/api/v1/vampire/update", srv.HostUrl))))
	srv.ShowMessClient(chatId, u.SUCCESS_ADDED_BOT)

	grl, _ := srv.As.GetAllGroupLinks()
	if len(grl) == 0 {
		return nil
	}
	err = srv.SendForceReply(chatId, fmt.Sprintf(u.GROUP_LINK_FOR_BOT_MSG, bot.Id))

	return err
}

func (srv *TgService) RM_delete_bot(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_delete_bot", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	id, err := strconv.Atoi(strings.TrimSpace(replyMes))
	if err != nil {
		srv.ShowMessClient(chatId, "неправильный формат id !")
		return err
	}
	bot, err := srv.As.GetBotInfoById(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			srv.ShowMessClient(chatId, "я не знаю такого бота !")
			return err
		}
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	if bot.IsDonor == 1 {
		srv.ShowMessClient(chatId, "главного бота нельзя удалить")
		return nil
	}
	// _, err = http.Get(fmt.Sprintf(srv.TgEndp, bot.Token, "setWebhook?url=")) // delete Webhook
	// if err != nil {
	// 	srv.l.Error("RM_delete_bot: delete Webhook", zap.Error(err))
	// }
	err = srv.As.DeleteBot(id)
	if err != nil {
		return err
	}
	err = srv.ShowMessClient(chatId, u.SUCCESS_DELETE_BOT)

	return err
}

func (srv *TgService) RM_add_ch_to_bot(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_add_ch_to_bot", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	id, err := strconv.Atoi(strings.TrimSpace(replyMes))
	if err != nil {
		srv.ShowMessClient(chatId, "неправильный формат id !")
		return err
	}
	bot, err := srv.As.GetBotInfoById(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			srv.ShowMessClient(chatId, "я не знаю такого бота !")
			return err
		}
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}

	err = srv.SendForceReply(chatId, fmt.Sprintf("укажите id канала в котором уже бот админ и к которому нужно привязать бота-%d", bot.Id))

	return err
}

func (srv *TgService) RM_add_ch_to_bot_spet2(m models.Update, botId int) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_add_ch_to_bot_spet2", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))
	replyMes = strings.TrimSpace(replyMes)

	chId, err := strconv.Atoi("-100"+replyMes)
	if err != nil {
		srv.ShowMessClient(chatId, fmt.Sprintf("%s: %v", u.ERR_MSG, err))
		return err
	}
	bot, err := srv.As.GetBotInfoById(botId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			srv.ShowMessClient(chatId, "я не знаю такого бота !")
			return err
		}
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}

	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chId),
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(
		fmt.Sprintf(srv.TgEndp, bot.Token, "getChat"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return fmt.Errorf("RM_add_ch_to_bot_spet2 POSt getChat err: %v", err)
	}
	defer resp.Body.Close()

	var j models.APIRBotresp
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return fmt.Errorf("RM_add_ch_to_bot_spet2 NewDecoder err: %v", err)
	}

	if !j.Ok {
		return fmt.Errorf("RM_add_ch_to_bot_spet2 !j.Ok error: %v. ch_id %d", j.Description, chId)
	}

	bot.ChId = j.Result.Id
	bot.ChLink = j.Result.InviteLink
	err = srv.As.EditBotChField(bot)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.ShowMessClient(chatId, fmt.Sprintf("канал %d привязанна к боту %d", chId, botId))
	return err
}

func (srv *TgService) RM_add_admin(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_add_admin", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	usr, err := srv.As.GetUserByUsername(strings.TrimSpace(replyMes))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			srv.ShowMessClient(chatId, "я не знаю такого юзера , пусть напишет мне /start")
			return err
		}
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.As.GetUserByUsername(%s) : %v", strings.TrimSpace(replyMes), err)
	}
	err = srv.As.EditAdmin(usr.Username, 1)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.As.EditAdmin(%s, 1) : %v", usr.Username, err)
	}
	err = srv.ShowMessClient(chatId, "Админ добавлен")
	return err
}

func (srv *TgService) RM_add_group_link(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_add_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	replyMes = strings.TrimSpace(replyMes)
	runeStr := []rune(replyMes)
	var groupLinkTitle string
	var groupLinkLink string
	for i := 0; i < len(runeStr); i++ {
		if i < 1 {
			continue
		}
		if string(runeStr[i-1]) == ":" && string(runeStr[i]) == ":" && string(runeStr[i+1]) == ":" {
			groupLinkTitle = string(runeStr[:i-1])
			groupLinkLink = string(runeStr[i+2:])
		}
	}

	// link := entity.NewGroupLink(groupLinkTitle, groupLinkLink)

	err := srv.As.AddNewGroupLink(groupLinkTitle, groupLinkLink)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.As.AddNewGroupLink(%s, %s) : %v", groupLinkTitle, groupLinkLink, err)
	}
	err = srv.ShowMessClient(chatId, "группа-ссылка добавлен")
	return err
}

func (srv *TgService) RM_edit_bot_group_link(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_edit_bot_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	replyMes = strings.TrimSpace(replyMes)
	runeStr := []rune(replyMes)
	var botIdStr string
	var groupLinkIdStr string
	for i := 0; i < len(runeStr); i++ {
		if i < 1 {
			continue
		}
		if string(runeStr[i-1]) == ":" && string(runeStr[i]) == ":" && string(runeStr[i+1]) == ":" {
			botIdStr = string(runeStr[:i-1])
			groupLinkIdStr = string(runeStr[i+2:])
		}
	}

	botId, err := strconv.Atoi(botIdStr)
	if err != nil {
		return fmt.Errorf("RM_edit_bot_group_link: некоректный id бота-%s : %v", botIdStr, err)
	}
	groupLinkId, err := strconv.Atoi(groupLinkIdStr)
	if err != nil {
		return fmt.Errorf("RM_edit_bot_group_link: некоректный id группы-ссылки-%s : %v", groupLinkIdStr, err)
	}

	bot, err := srv.As.GetBotInfoById(botId)
	if err != nil {
		return fmt.Errorf("RM_edit_bot_group_link: GetBotInfoById-%d : %v", botId, err)
	}
	oldGroupLink := bot.GroupLinkId

	err = srv.As.EditBotGroupLinkId(groupLinkId, botId)
	if err != nil {
		return fmt.Errorf("RM_edit_bot_group_link: EditBotGroupLinkId-%d grId-%d : %v", botId, groupLinkId, err)
	}
	
	err = srv.ShowMessClient(chatId, fmt.Sprintf("для бота %d, ссылка успешно изменена %d -> %d", botId, oldGroupLink, groupLinkId))
	return err
}

func (srv *TgService) RM_delete_group_link(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_delete_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	replyMes = strings.TrimSpace(replyMes)
	grId, err := strconv.Atoi(replyMes)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.As.DeleteGroupLink(grId)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.As.EditBotGroupLinkIdToNull(grId)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.ShowMessClient(chatId, "группа-ссылка удалена")
	return err
}

func (srv *TgService) RM_update_bot_group_link(m models.Update, botId int) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_update_bot_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))
	replyMes = strings.TrimSpace(replyMes)

	grId, err := strconv.Atoi(replyMes)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.As.EditBotGroupLinkId(grId, botId)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.ShowMessClient(chatId, fmt.Sprintf("группа-ссылка %d привязанна к боту %d", grId, botId))
	return err
}

func (srv *TgService) RM_update_group_link(m models.Update, refId int) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_update_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))
	replyMes = strings.TrimSpace(replyMes)

	err := srv.As.UpdateGroupLink(refId, replyMes)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.ShowMessClient(chatId, "группа-ссылка обновлена")
	return err
}
