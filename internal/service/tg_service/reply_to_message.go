package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myapp/internal/models"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

const (
	ERR_MSG            = "что то пошло не так, попробуйте позже"
	ERR_MSG_2          = "что то пошло не так, попробуйте еще: "
	SUCCESS_DELETE_BOT = "бот успешно удален"
	SUCCESS_ADDED_BOT  = "бот успешно создан"
)

const (
	NEW_ADMIN_MSG           = "Укажите username нового админа без '@' пример: adminchik0 вместо @adminchik"
	NEW_BOT_MSG             = "Укажите токен нового бота:"
	DELETE_BOT_MSG          = "Укажите id бота которого нужно удалить:"
	ADD_CH_TO_BOT_MSG       = "Укажите id бота для которого нужно добавить канал:"
	NEW_GROUP_LINK_MSG      = "Укажите название новой группы-ссылки и саму ссылку которую подставлять в таком формате -> моя группа 1:::ya.ru"
	EDIT_BOT_GROUP_LINK_MSG = "Укажите токен бота для которого нужно поменять группу-ссылку"
	DELETE_GROUP_LINK_MSG   = "Укажите id группы-ссылки которого нужно удалить:"
	UPDATE_GROUP_LINK_MSG   = "Укажите id группы-ссылки которую нужно поменять:"
	GROUP_LINK_FOR_BOT_MSG  = "укажите номер группы-ссылки для нового бота[%d"
	GROUP_LINK_UPDATE_MSG   = "укажите новую ссылку для ref [%d"

	DELETE_POST_MSG = "Укажите id поста в доноре"
)


func (srv *TgService) HandleReplyToMessage(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	// chatId := m.Message.From.Id

	srv.l.Info("tgClient: HandleReplyToMessage", zap.Any("rm.Tex", rm.Text), zap.Any("replyMes", replyMes))

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
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_obtain_vampire_bot_token", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	tgobotResp, err := srv.getBotByToken(strings.TrimSpace(replyMes))
	if err != nil {
		return err
	}
	res := tgobotResp.Result
	err = srv.db.AddNewBot(res.Id, res.UserName, res.FirstName, replyMes, 0)
	if err != nil {
		return err
	}
	srv.SendMessage(chatId, SUCCESS_ADDED_BOT)

	grl, _ := srv.db.GetAllGroupLinks()
	if len(grl) == 0 {
		return nil
	}
	err = srv.SendForceReply(chatId, fmt.Sprintf(GROUP_LINK_FOR_BOT_MSG, res.Id))

	return err
}

func (srv *TgService) RM_delete_bot(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_delete_bot", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	id, err := strconv.Atoi(strings.TrimSpace(replyMes))
	if err != nil {
		srv.SendMessage(chatId, "неправильный формат id !")
		return err
	}
	bot, err := srv.db.GetBotInfoById(id)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	if bot.Id == 0 {
		srv.SendMessage(chatId, "я не знаю такого бота !")
		return nil
	}
	if bot.IsDonor == 1 {
		srv.SendMessage(chatId, "главного бота нельзя удалить")
		return nil
	}
	err = srv.db.DeleteBot(id)
	if err != nil {
		return err
	}
	err = srv.SendMessage(chatId, SUCCESS_DELETE_BOT)

	return err
}

func (srv *TgService) RM_add_ch_to_bot(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_add_ch_to_bot", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	id, err := strconv.Atoi(strings.TrimSpace(replyMes))
	if err != nil {
		srv.SendMessage(chatId, "неправильный формат id !")
		return err
	}
	bot, err := srv.db.GetBotInfoById(id)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	if bot.Id == 0 {
		srv.SendMessage(chatId, "я не знаю такого бота !")
		return nil
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

	chId, err := strconv.Atoi("-100" + replyMes)
	if err != nil {
		srv.SendMessage(chatId, fmt.Sprintf("%s: %v", ERR_MSG, err))
		return err
	}
	bot, err := srv.db.GetBotInfoById(botId)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	if bot.Id == 0 {
		srv.SendMessage(chatId, "я не знаю такого бота !")
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

	var j models.APIRBotresp
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return fmt.Errorf("RM_add_ch_to_bot_spet2 NewDecoder err: %v", err)
	}

	if !j.Ok {
		return fmt.Errorf("RM_add_ch_to_bot_spet2 !j.Ok error: %v. ch_id %d", j.Description, chId)
	}

	bot.ChId = j.Result.Id
	bot.ChLink = j.Result.InviteLink
	err = srv.db.EditBotField(bot.Id, "ch_id", bot.ChId)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	err = srv.db.EditBotField(bot.Id, "ch_link", bot.ChLink)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	err = srv.SendMessage(chatId, fmt.Sprintf("канал %d привязанна к боту %d", chId, botId))
	return err
}

func (srv *TgService) RM_add_admin(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_add_admin", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	usr, err := srv.db.GetUserByUsername(strings.TrimSpace(replyMes))
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.db.GetUserByUsername(%s) : %v", strings.TrimSpace(replyMes), err)
	}
	if usr.Id == 0 {
		srv.SendMessage(chatId, "я не знаю такого юзера , пусть напишет мне /start")
		return nil
	}
	err = srv.db.EditAdmin(usr.Username, 1)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.db.EditAdmin(%s, 1) : %v", usr.Username, err)
	}
	err = srv.SendMessage(chatId, "Админ добавлен")
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

			groupLinkTitle = strings.TrimSpace(groupLinkTitle)
			groupLinkLink = strings.TrimSpace(groupLinkLink)
		}
	}

	err := srv.db.AddNewGroupLink(groupLinkTitle, groupLinkLink)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.db.AddNewGroupLink(%s, %s) : %v", groupLinkTitle, groupLinkLink, err)
	}
	err = srv.SendMessage(chatId, "группа-ссылка добавлен")
	return err
}

func (srv *TgService) RM_edit_bot_group_link(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	srv.l.Info("tg_service: RM_edit_bot_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	botToken := replyMes
	urlArr := strings.Split(botToken, ":")
	if len(urlArr) != 2 {
		return fmt.Errorf("RM_edit_bot_group_link err: не правилный токен %s", botToken)
	}
	botIdStr := urlArr[0]

	_, err := strconv.Atoi(botIdStr)
	if err != nil {
		return fmt.Errorf("RM_edit_bot_group_link: некоректный id бота-%s : %v", botIdStr, err)
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
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_delete_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	replyMes = strings.TrimSpace(replyMes)
	grId, err := strconv.Atoi(replyMes)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	err = srv.db.DeleteGroupLink(grId)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	err = srv.db.EditBotGroupLinkIdToNull(grId)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	err = srv.SendMessage(chatId, "группа-ссылка удалена")
	return err
}

func (srv *TgService) RM_delete_post_in_chs(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	srv.l.Info("tg_service: RM_delete_post_in_chs", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	linkToPostInCh := replyMes
	chIdStrFromLink, postIdStrFromLink, err := srv.GetPostAndChFromLonk(linkToPostInCh)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs err: %v", err)
	}
	postIdFromLink, err := strconv.Atoi(postIdStrFromLink)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs Atoi postIdStrFromLink-%s err: %v", postIdStrFromLink, err)
	}
	chIdFromLink, err := strconv.Atoi(chIdStrFromLink)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs Atoi chIdStrFromLink-%s err: %v", chIdStrFromLink, err)
	}

	chwithmStr := fmt.Sprintf("-100%d", chIdFromLink)
	chDonor, err := strconv.Atoi(chwithmStr)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs Atoi-%s err: %v", chwithmStr, err)
	}

	bot, err := srv.db.GetBotByChannelId(chDonor)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs GetBotByChannelId chDonor-%d err: %v", chDonor, err)
	}
	if bot.IsDonor == 0 {
		return fmt.Errorf("RM_delete_post_in_chs канал не донор chDonor-%d err: bot.IsDonor == 0", chDonor)
	}

	newm := models.Update{
		EditedChannelPost: &models.Message{
			Chat: &models.Chat{},
		},
	}
	m.EditedChannelPost.Chat.Id = fromId
	m.EditedChannelPost.MessageId = postIdFromLink
	m.EditedChannelPost.Text = "deletepost"
	err = srv.Donor_HandleEditedChannelPost(newm)
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chsDonor_HandleEditedChannelPost newm-%+v err: %v", newm, err)
	}
	srv.SendMessage(fromId, "пост удален")
	return nil
}

func (srv *TgService) RM_update_bot_group_link(m models.Update, botId int) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_update_bot_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))
	replyMes = strings.TrimSpace(replyMes)

	grId, err := strconv.Atoi(replyMes)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	err = srv.db.EditBotGroupLinkId(grId, botId)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	err = srv.SendMessage(chatId, fmt.Sprintf("группа-ссылка %d привязанна к боту %d", grId, botId))
	return err
}

func (srv *TgService) RM_update_group_link(m models.Update, refId int) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_update_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))
	replyMes = strings.TrimSpace(replyMes)

	err := srv.db.UpdateGroupLink(refId, replyMes)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG)
		return err
	}
	err = srv.SendMessage(chatId, "группа-ссылка обновлена")
	return err
}
