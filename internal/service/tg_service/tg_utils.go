package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myapp/internal/models"
	"net/http"
	"strconv"
)

func (srv *TgService) showAdminPanel(chatId int) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
		"text":    "Привет, я бот Донор",
		"reply_markup": `{"inline_keyboard" : [
			[{ "text": "Привязанные боты и каналы", "callback_data": "show_bots_and_channels" }],
			[{ "text": "Добавить бота", "callback_data": "create_vampere_bot" }],
			[{ "text": "Удалить бота", "callback_data": "delete_vampere_bot" }],
			[{ "text": "Добавить канал боту", "callback_data": "add_ch_to_bot" }],
			[{ "text": "Добавить группу-ссылку", "callback_data": "create_group_link" }],
			[{ "text": "Удалить группу-ссылку", "callback_data": "delete_group_link" }],
			[{ "text": "Редактировать группу-ссылку", "callback_data": "update_group_link" }],
			[{ "text": "Поменять группу-ссылку у бота", "callback_data": "edit_bot_group_link" }],
			[{ "text": "Все группы-ссылки", "callback_data": "show_all_group_links" }],
			[{ "text": "Добавить Админа", "callback_data": "add_admin_btn" }],
			[{ "text": "Удалить потеряных ботов", "callback_data": "del_lost_bots" }],
			[{ "text": "Restart app", "callback_data": "restart_app" }]
		]}`,
	})
	if err != nil {
		return err
	}
	err = srv.sendData(json_data)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) getBotByToken(token string) (models.APIRBotresp, error) {
	resp, err := http.Get(fmt.Sprintf(
		srv.TgEndp, token, "getMe",
	))
	if err != nil {
		return models.APIRBotresp{}, err
	}
	defer resp.Body.Close()

	var j models.APIRBotresp
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return models.APIRBotresp{}, err
	}
	return j, err
}

func (srv *TgService) showBotsAndChannels(chatId int) error {
	bots, err := srv.As.GetAllBots()
	if err != nil {
		return err
	}
	var mess bytes.Buffer
	for i, b := range bots {
		mess.WriteString(fmt.Sprintf("%d) id: %d - @%s ", i+1, b.Id, b.Username))
		if b.IsDonor == 1 {
			mess.WriteString("-Донор")
		}
		mess.WriteString(fmt.Sprintf("\n	ch_link: %s\n", b.ChLink))

		if i % 50 == 0 && i > 0 {
			json_data, err := json.Marshal(map[string]any{
				"chat_id": strconv.Itoa(chatId),
				"text":    mess.String(),
				"reply_markup": `{"inline_keyboard" : [
					[{ "text": "Назад", "callback_data": "show_admin_panel" }]
				]}`,
			})
			if err != nil {
				return err
			}
			err = srv.sendData(json_data)
			if err != nil {
				return err
			}
			mess.Reset()
		}
	}
	txt := mess.String()
	if len(txt) > 4000 {
		txt = txt[:4000]
	}
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
		"text":    txt,
		"reply_markup": `{"inline_keyboard" : [
			[{ "text": "Назад", "callback_data": "show_admin_panel" }]
		]}`,
	})
	if err != nil {
		return err
	}
	err = srv.sendData(json_data)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) showAllGroupLinks(chatId int) error {
	grs, err := srv.As.GetAllGroupLinks()
	if err != nil {
		return err
	}
	var mess bytes.Buffer
	for i, b := range grs {
		mess.WriteString(fmt.Sprintf("%d) id: %d\n", i+1, b.Id))
		mess.WriteString(fmt.Sprintf("Название: %s\n", b.Title))
		mess.WriteString(fmt.Sprintf("Ссылка: %s\n", b.Link))
		bots, err := srv.As.GetBotsByGrouLinkId(b.Id)
		if err != nil {
			return err
		}
		mess.WriteString(fmt.Sprintf("Количество Привязаных ботов: %d\n", len(bots)))
		mess.WriteString("\n")
	}
	txt := mess.String()
	if len(txt) > 4000 {
		txt = txt[:4000]
	}
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
		"text":    txt,
		"reply_markup": `{"inline_keyboard" : [ 
			[{ "text": "Назад", "callback_data": "show_admin_panel" }]
		]}`,
	})
	if err != nil {
		return err
	}
	err = srv.sendData(json_data)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) getChatByCurrBot(chatId int, token string) (models.GetChatResult, error) {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
	})
	if err != nil {
		return models.GetChatResult{}, err
	}
	resp, err := http.Post(
		fmt.Sprintf(srv.TgEndp, token, "getChat"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return models.GetChatResult{}, err
	}
	defer resp.Body.Close()
	var cAny models.GetChatResult
	if err := json.NewDecoder(resp.Body).Decode(&cAny); err != nil {
		return models.GetChatResult{}, err
	}
	return cAny, nil
}

func (srv *TgService) SendForceReply(chat int, mess string) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id":      strconv.Itoa(chat),
		"text":         mess,
		"reply_markup": `{"force_reply": true}`,
	})
	if err != nil {
		return err
	}
	err = srv.sendData(json_data)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) ShowMessClient(chat int, mess string) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chat),
		"text":    mess,
	})
	if err != nil {
		return err
	}
	err = srv.sendData(json_data)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) DeleteMess(chat, messId int) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id":    strconv.Itoa(chat),
		"message_id": strconv.Itoa(messId),
	})
	if err != nil {
		return err
	}
	_, err = http.Post(
		fmt.Sprintf(srv.TgEndp, srv.Token, "deleteMessage"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) sendData(json_data []byte) error {
	_, err := http.Post(
		fmt.Sprintf(srv.TgEndp, srv.Token, "sendMessage"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return err
	}
	return nil
}
