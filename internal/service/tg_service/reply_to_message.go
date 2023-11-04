package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myapp/internal/models"
	"net/http"
	"strconv"
	"strings"
)


func (srv *TgService) HandleReplyToMessage(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("HandleCallbackQuery: fromId: %d fromUsername: %s, replyMes: %s rm.Tex: %s", fromId, fromUsername, replyMes, rm.Text))

	if rm.Text == NEW_BOT_MSG {
		err := srv.RM_obtain_vampire_bot_token(m)
		return err
	}

	if rm.Text == DELETE_BOT_MSG {
		err := srv.RM_delete_bot(m)
		return err
	}

	if rm.Text == ADD_CH_TO_BOT_MSG {
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

	if rm.Text == NEW_ADMIN_MSG {
		err := srv.RM_add_admin(m)
		return err
	}

	if rm.Text == NEW_GROUP_LINK_MSG {
		err := srv.RM_add_group_link(m)
		return err
	}

	if rm.Text == EDIT_BOT_GROUP_LINK_MSG {
		err := srv.RM_edit_bot_group_link(m)
		return err
	}

	if rm.Text == DELETE_GROUP_LINK_MSG {
		err := srv.RM_delete_group_link(m)
		return err
	}

	if rm.Text == DELETE_POST_MSG {
		err := srv.RM_delete_post_in_chs(m)
		return err
	}

	if rm.Text == UPDATE_GROUP_LINK_MSG {
		chatId := m.Message.From.Id
		replyMes := m.Message.Text
		replyMes = strings.TrimSpace(replyMes)

		grId, err := strconv.Atoi(replyMes)
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG)
			return err
		}
		err = srv.SendForceReply(chatId, fmt.Sprintf(GROUP_LINK_UPDATE_MSG, grId))
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
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_obtain_vampire_bot_token: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	tgobotResp, err := srv.GetMe(replyMes)
	if err != nil {
		return err
	}
	res := tgobotResp.Result
	err = srv.db.AddNewBot(res.Id, res.UserName, res.FirstName, replyMes, 0)
	if err != nil {
		return err
	}
	srv.SendMessage(fromId, SUCCESS_ADDED_BOT)

	grl, _ := srv.db.GetAllGroupLinks()
	if len(grl) == 0 {
		return nil
	}
	srv.SendForceReply(fromId, fmt.Sprintf(GROUP_LINK_FOR_BOT_MSG, res.Id))

	return nil
}

func (srv *TgService) RM_delete_bot(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_delete_bot: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	id, err := strconv.Atoi(strings.TrimSpace(replyMes))
	if err != nil {
		srv.SendMessage(fromId, "неправильный формат id !")
		return err
	}
	bot, err := srv.db.GetBotInfoById(id)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	if bot.Id == 0 {
		srv.SendMessage(fromId, "я не знаю такого бота !")
		return nil
	}
	if bot.IsDonor == 1 {
		srv.SendMessage(fromId, "главного бота нельзя удалить")
		return nil
	}
	err = srv.db.DeleteBot(id)
	if err != nil {
		return err
	}
	srv.SendMessage(fromId, SUCCESS_DELETE_BOT)

	return nil
}

func (srv *TgService) RM_add_ch_to_bot(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_add_ch_to_bot: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	id, err := strconv.Atoi(strings.TrimSpace(replyMes))
	if err != nil {
		srv.SendMessage(fromId, "неправильный формат id !")
		return err
	}
	bot, err := srv.db.GetBotInfoById(id)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	if bot.Id == 0 {
		srv.SendMessage(fromId, "я не знаю такого бота !")
		return nil
	}

	srv.SendForceReply(fromId, fmt.Sprintf("укажите id канала в котором уже бот админ и к которому нужно привязать бота: %d", bot.Id))

	return nil
}

func (srv *TgService) RM_add_ch_to_bot_spet2(m models.Update, botId int) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_add_ch_to_bot_spet2: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))
	replyMes = strings.TrimSpace(replyMes)

	chId, err := strconv.Atoi("-100" + replyMes)
	if err != nil {
		srv.SendMessage(fromId, fmt.Sprintf("%s: %v", ERR_MSG, err))
		return err
	}
	bot, err := srv.db.GetBotInfoById(botId)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	if bot.Id == 0 {
		srv.SendMessage(fromId, "я не знаю такого бота !")
		return nil
	}

	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chId),
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgEndp, bot.Token, "getChat"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return fmt.Errorf("RM_add_ch_to_bot_spet2 POSt getChat err: %v", err)
	}
	defer resp.Body.Close()

	var j models.ApiBotResp
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return fmt.Errorf("RM_add_ch_to_bot_spet2 NewDecoder err: %v", err)
	}

	if !j.Ok {
		return fmt.Errorf("RM_add_ch_to_bot_spet2 !j.Ok error: %v, ch_id %d", j.Description, chId)
	}

	bot.ChId = j.Result.Id
	bot.ChLink = j.Result.InviteLink
	err = srv.db.EditBotField(bot.Id, "ch_id", bot.ChId)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	err = srv.db.EditBotField(bot.Id, "ch_link", bot.ChLink)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	srv.SendMessage(fromId, fmt.Sprintf("канал %d привязанна к боту %d", chId, botId))
	return nil
}

func (srv *TgService) RM_add_admin(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_add_admin: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	usr, err := srv.db.GetUserByUsername(strings.TrimSpace(replyMes))
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.db.GetUserByUsername(%s) : %v", strings.TrimSpace(replyMes), err)
	}
	if usr.Id == 0 {
		srv.SendMessage(fromId, "я не знаю такого юзера , пусть напишет мне /start")
		return nil
	}
	err = srv.db.EditAdmin(usr.Username, 1)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.db.EditAdmin(%s, 1) err: %v", usr.Username, err)
	}
	srv.SendMessage(fromId, "Админ добавлен")
	return nil
}

func (srv *TgService) RM_add_group_link(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_add_group_link: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

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

			groupLinkTitle = strings.TrimSpace(groupLinkTitle)
			groupLinkLink = strings.TrimSpace(groupLinkLink)
		}
	}

	err := srv.db.AddNewGroupLink(groupLinkTitle, groupLinkLink)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.db.AddNewGroupLink(%s, %s) err: %v", groupLinkTitle, groupLinkLink, err)
	}
	srv.SendMessage(fromId, "группа-ссылка добавлен")
	return nil
}

func (srv *TgService) RM_edit_bot_group_link(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_edit_bot_group_link: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	botToken := replyMes
	urlArr := strings.Split(botToken, ":")
	if len(urlArr) != 2 {
		return fmt.Errorf("RM_edit_bot_group_link err: не правилный токен %s", botToken)
	}
	botIdStr := urlArr[0]

	_, err := strconv.Atoi(botIdStr)
	if err != nil {
		return fmt.Errorf("RM_edit_bot_group_link: некоректный id бота: %s err: %v", botIdStr, err)
	}

	grs, _ := srv.db.GetAllGroupLinks()
	var mess bytes.Buffer
	mess.WriteString(`{"inline_keyboard" : [`)
	for _, v := range grs {
		mess.WriteString(fmt.Sprintf(`[{ "text": "%s", "callback_data": "edit_bot_%s_link_to_%d_gr_link_btn" }],`, v.Title, botIdStr, v.Id))
	}
	mess.WriteString(`[{ "text": "назад", "callback_data": "_todo_" }]`)
	mess.WriteString(`]}`)
	json_data, err := json.Marshal(map[string]any{
		"chat_id":      strconv.Itoa(fromId),
		"text":         "выберете группу-ссылку",
		"reply_markup": mess.String(),
	})
	if err != nil {
		return fmt.Errorf("RM__DEL_ADV_POST Marshal err %v", err)
	}
	err = srv.sendData(json_data, "sendMessage")
	if err != nil {
		return err
	}

	// err = srv.SendMessage(chatId, fmt.Sprintf("для бота %d, ссылка успешно изменена %d -> %d", botId, oldGroupLink, groupLinkId))
	return nil
}

func (srv *TgService) RM_delete_group_link(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_delete_group_link: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	replyMes = strings.TrimSpace(replyMes)
	grId, err := strconv.Atoi(replyMes)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	err = srv.db.DeleteGroupLink(grId)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	err = srv.db.EditBotGroupLinkIdToNull(grId)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	srv.SendMessage(fromId, "группа-ссылка удалена")
	return nil
}

func (srv *TgService) RM_delete_post_in_chs(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_delete_post_in_chs: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	linkToPostInCh := replyMes
	chIdStrFromLink, postIdStrFromLink, err := srv.GetPostAndChFromLink(linkToPostInCh)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs err: %v", err)
	}
	postIdFromLink, err := strconv.Atoi(postIdStrFromLink)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs Atoi postIdStrFromLink: %s err: %v", postIdStrFromLink, err)
	}
	chIdFromLink, err := strconv.Atoi(chIdStrFromLink)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs Atoi chIdStrFromLink: %s err: %v", chIdStrFromLink, err)
	}

	chwithmStr := fmt.Sprintf("-100%d", chIdFromLink)
	chDonor, err := strconv.Atoi(chwithmStr)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs Atoi: %s err: %v", chwithmStr, err)
	}

	bot, err := srv.db.GetBotByChannelId(chDonor)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs GetBotByChannelId chDonor: %d err: %v", chDonor, err)
	}
	if bot.IsDonor == 0 {
		return fmt.Errorf("RM_delete_post_in_chs канал не донор chDonor: %d err: bot.IsDonor == 0", chDonor)
	}
	srv.l.Info("10")
	newm := models.Update{
		EditedChannelPost: &models.Message{
			Chat: &models.Chat{},
		},
	}
	srv.l.Info("11")
	newm.EditedChannelPost.Chat.Id = fromId
	srv.l.Info("12")
	newm.EditedChannelPost.MessageId = postIdFromLink
	srv.l.Info("13")
	cap := "deletepost"
	newm.EditedChannelPost.Text = cap
	srv.l.Info("14")
	newm.EditedChannelPost.Caption = &cap
	srv.l.Info("15")
	newm.EditedChannelPost.Video = &models.Video{}
	srv.l.Info("16")
	
	// if len(m.EditedChannelPost.Photo) > 0 {
	// 	srv.l.Info("PhotoPhoto")
	// 	newm.EditedChannelPost.Photo = make([]models.PhotoSize, 1, 1)
	// 	srv.l.Info("17")
	// } else if m.EditedChannelPost.Video != nil {
	// 	srv.l.Info("VideoVideo")
	// 	newm.EditedChannelPost.Video = &models.Video{}
	// 	srv.l.Info("16")
	// }
	err = srv.Donor_HandleEditedChannelPost(newm)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chsDonor_HandleEditedChannelPost newm: %+v err: %v", newm, err)
	}
	srv.SendMessage(fromId, "пост удален")
	return nil
}

func (srv *TgService) RM_update_bot_group_link(m models.Update, botId int) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_update_bot_group_link: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))
	replyMes = strings.TrimSpace(replyMes)

	grId, err := strconv.Atoi(replyMes)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	err = srv.db.EditBotGroupLinkId(grId, botId)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	srv.SendMessage(fromId, fmt.Sprintf("группа-ссылка %d привязанна к боту %d", grId, botId))
	return nil
}

func (srv *TgService) RM_update_group_link(m models.Update, refId int) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_update_group_link: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))
	replyMes = strings.TrimSpace(replyMes)

	err := srv.db.UpdateGroupLink(refId, replyMes)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	srv.SendMessage(fromId, "группа-ссылка обновлена")
	return nil
}
