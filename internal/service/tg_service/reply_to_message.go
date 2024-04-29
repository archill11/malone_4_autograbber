package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myapp/internal/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)


func (srv *TgService) HandleReplyToMessage(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("HandleReplyToMessage: fromId: %d fromUsername: %s, replyMes: %s rm.Tex: %s", fromId, fromUsername, replyMes, rm.Text))

	if rm.Text == NEW_BOT_MSG {
		err := srv.RM_obtain_vampire_bot_token(m)
		return err
	}

	if rm.Text == DELETE_BOT_MSG {
		err := srv.RM_delete_bot(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == ADD_CH_TO_BOT_MSG {
		err := srv.RM_add_ch_to_bot(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if strings.HasPrefix(rm.Text, "укажите id канала в котором уже бот админ и к которому нужно привязать бота: ") {
		runes := []rune(rm.Text)
		runesStr := string(runes[len([]rune("укажите id канала в котором уже бот админ и к которому нужно привязать бота: ")):])
		botId, _ := strconv.Atoi(runesStr)
		err := srv.RM_add_ch_to_bot_spet2(m, botId)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == NEW_GROUP_LINK_MSG {
		err := srv.RM_add_group_link(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == EDIT_BOT_GROUP_LINK_MSG {
		err := srv.RM_edit_bot_group_link(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == EDIT_BOT_LICHKA_MSG {
		err := srv.RM_edit_bot_lichka(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == EDIT_BOT_LICHKA_BY_GRLINK_MSG {
		err := srv.RM_edit_bot_lichka_by_gr_link(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == DELETE_GROUP_LINK_MSG {
		err := srv.RM_delete_group_link(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == DELETE_POST_MSG {
		err := srv.RM_delete_post_in_chs(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == NEW_ADMIN_MSG {
		err := srv.RM__NEW_ADMIN_MSG(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == DEL_ADMIN_MSG {
		err := srv.RM__DEL_ADMIN_MSG(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == NEW_USER_MSG {
		err := srv.RM__NEW_USER_MSG(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == DEL_USER_MSG {
		err := srv.RM__DEL_USER_MSG(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == CHANGE_DOMEN_MSG {
		err := srv.RM__CHANGE_DOMEN_MSG(m)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if rm.Text == UPDATE_GROUP_LINK_MSG {
		chatId := m.Message.From.Id
		replyMes := m.Message.Text
		replyMes = strings.TrimSpace(replyMes)

		grId, err := strconv.Atoi(replyMes)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
			return err
		}
		err = srv.SendForceReply(chatId, fmt.Sprintf(GROUP_LINK_UPDATE_MSG, grId))
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if strings.HasPrefix(rm.Text, "укажите номер группы-ссылки для нового бота[") {
		runes := []rune(rm.Text)
		runesStr := string(runes[len([]rune("укажите номер группы-ссылки для нового бота[")):])
		botId, _ := strconv.Atoi(runesStr)
		err := srv.RM_update_bot_group_link(m, botId)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
		return err
	}

	if strings.HasPrefix(rm.Text, "укажите новую ссылку для ref [") {
		runes := []rune(rm.Text)
		runesStr := string(runes[len([]rune("укажите новую ссылку для ref [")):])
		refId, _ := strconv.Atoi(runesStr)
		err := srv.RM_update_group_link(m, refId)
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG)
			srv.SendMessage(fromId, err.Error())
		}
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
	srv.db.EditBotUserCreator(res.Id, fromId)

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

func (srv *TgService) RM_edit_bot_lichka(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_edit_bot_lichka: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	words := strings.Fields(replyMes)
	if len(words) != 2 {
		return fmt.Errorf("неверный формат ввода")
	}

	botToken := words[0]
	urlArr := strings.Split(botToken, ":")
	botIdStr := urlArr[0]
	botId, err := strconv.Atoi(botIdStr)
	if err != nil {
		return fmt.Errorf("RM_edit_bot_lichka: некоректный id бота: %s err: %v", botIdStr, err)
	}

	lichka := words[1]
	if lichka != "" {
		lichka = srv.AddAt(words[1])
	}
	err = srv.db.EditBotLichka(botId, lichka)
	if err != nil {
		return fmt.Errorf("RM_edit_bot_lichka: EditBotLichka err: %v", err)
	}

	srv.SendMessage(fromId, fmt.Sprintf("для бота %d, личка успешно изменена -> %s", botId, lichka))
	return nil
}

func (srv *TgService) RM_edit_bot_lichka_by_gr_link(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM_edit_bot_lichka_by_gr_link: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	words := strings.Fields(replyMes)
	if len(words) < 2 {
		return fmt.Errorf("неверный формат ввода")
	}
	lichka := words[0]
	if lichka != "" {
		lichka = srv.AddAt(lichka)
	}

	for i, v := range words {
		if i == 0 {
			continue
		}
		rgLinkId, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("RM_edit_bot_lichka_by_gr_link: некоректный id группы ссылки: %s err: %v", v, err)
		}
		bots, err := srv.db.GetBotsByGrouLinkId(rgLinkId)
		if err != nil {
			return fmt.Errorf("RM_edit_bot_lichka_by_gr_link: GetBotsByGrouLinkId err: %v", err)
		}
		for _, vv := range bots {
			srv.db.EditBotLichka(vv.Id, lichka)
		}
		srv.SendMessage(fromId, fmt.Sprintf("для гр-ссылки %d (кол-во ботов: %d), личка успешно изменена -> %s", rgLinkId, len(bots), lichka))
	}

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

	allVampBots, err := srv.db.GetAllVampBots()
	if err != nil {
		return fmt.Errorf("RM_delete_post_in_chs GetAllVampBots err : %v", err)
	}
	for i, vampBot := range allVampBots {
		if vampBot.ChId == 0 {
			continue
		}
		err := srv.deleteChPostAsVamp(vampBot, m, postIdFromLink)
		if err != nil {
			srv.l.Error("RM_delete_post_in_chs: deleteChPostAsVamp err", zap.Error(err))
		}
		srv.l.Info("RM_delete_post_in_chs", zap.Any("bot index in arr", i), zap.Any("bot ch link", vampBot.ChLink))
		time.Sleep(time.Second)
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
	groupLink, _ := srv.db.GetGroupLinkById(grId)
	if groupLink.Id == 0 {
		srv.SendMessage(fromId, ERR_MSG)
		srv.SendMessage(fromId, fmt.Sprintf("нету такой группы-ссылки: %v", grId))
		return nil
	}

	err = srv.db.EditBotGroupLinkId(grId, botId)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return err
	}
	srv.SendMessage(fromId, fmt.Sprintf("группа-ссылка %d привязанна к боту %d", grId, botId))

	// srv.SendForceReply(fromId, EDIT_BOT_LICHKA_MSG)
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

func (srv *TgService) RM__NEW_ADMIN_MSG(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM__NEW_ADMIN_MSG: fromId-%d fromUsername-%s, replyMes-%s", fromId, fromUsername, replyMes))

	username := replyMes
	username = srv.DelAt(username)

	usr, err := srv.db.GetUserByUsername(username)
	if err != nil {
		srv.SendMessage(fromId, ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.db.GetUserByUsername(%s) : %v", strings.TrimSpace(replyMes), err)
	}
	if usr.Id == 0 {
		srv.SendMessage(fromId, "я не знаю такого юзера , пусть напишет мне /start")
		return nil
	}

	err = srv.db.EditAdmin(username, 1)
	if err != nil {
		return err
	}
	srv.SendMessage(fromId, "админ добавлен успешно")
	return nil
}

func (srv *TgService) RM__DEL_ADMIN_MSG(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM__DEL_ADMIN_MSG: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	username := replyMes
	username = srv.DelAt(username)

	err := srv.db.EditAdmin(username, 0)
	if err != nil {
		return err
	}
	srv.SendMessage(fromId, "админ удален успешно")
	return nil
}

func (srv *TgService) RM__NEW_USER_MSG(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM__NEW_USER_MSG: fromId-%d fromUsername-%s, replyMes-%s", fromId, fromUsername, replyMes))

	username := replyMes
	username = srv.DelAt(username)

	err := srv.db.EditIsUser(username, 1)
	if err != nil {
		return err
	}
	srv.SendMessage(fromId, "user добавлен успешно")
	return nil
}

func (srv *TgService) RM__DEL_USER_MSG(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM__DEL_USER_MSG: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	username := replyMes
	username = srv.DelAt(username)

	err := srv.db.EditIsUser(username, 0)
	if err != nil {
		return err
	}
	srv.SendMessage(fromId, "user удален успешно")
	return nil
}

func (srv *TgService) RM__CHANGE_DOMEN_MSG(m models.Update) error {
	replyMes := m.Message.Text
	fromId := m.Message.From.Id
	fromUsername := m.Message.From.UserName
	srv.l.Info(fmt.Sprintf("RM__CHANGE_DOMEN_MSG: fromId: %d fromUsername: %s, replyMes: %s", fromId, fromUsername, replyMes))

	words := strings.Fields(replyMes)
	if len(words) < 2 {
		return fmt.Errorf("неверный формат ввода")
	}
	old_domen := words[0]
	new_domen := words[1]

	allgr, err := srv.db.GetAllGroupLinks()
	if err != nil {
		return fmt.Errorf("RM__CHANGE_DOMEN_MSG GetAllGroupLinks err: %v", err)
	}
	var cnt int
	for _, v := range allgr {
		oldLink := v.Link
		newLink := strings.Replace(oldLink, old_domen, new_domen, -1)
		if oldLink == newLink {
			continue
		}
		err = srv.db.UpdateGroupLink(v.Id, newLink)
		if err != nil {
			err = fmt.Errorf("RM__CHANGE_DOMEN_MSG UpdateGroupLink err: %v", err)
			srv.SendMessage(fromId, err.Error())
		}
		cnt++
	}

	logMess := fmt.Sprintf("все ссылки изменены. %d шт", cnt)
	srv.SendMessage(fromId, logMess)

	return nil
}