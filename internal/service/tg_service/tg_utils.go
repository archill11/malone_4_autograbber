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
			[{ "text": "Привязанные боты и  каналы", "callback_data": "show_bots_and_channels" }],
			[{ "text": "Добавить бота", "callback_data": "create_vampere_bot" }],
			[{ "text": "Удалить бота", "callback_data": "delete_vampere_bot" }],
			[{ "text": "Добавить Админа", "callback_data": "add_admin_btn" }]
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
	resp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, token, "getMe"),
	)
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
	var isDonor bool
	for i, b := range bots {
		if b.IsDonor == 1 {
			isDonor = true
		} else {
			isDonor = false
		}
		mess.WriteString(fmt.Sprintf("%d) id: %d\n", i+1, b.Id))
		mess.WriteString(fmt.Sprintf("@%s\n", b.Username))
		mess.WriteString(fmt.Sprintf("Донор: %t\n", isDonor))
		mess.WriteString("Привязанный канал:\n")
		mess.WriteString(fmt.Sprintf("id: %d\n", b.ChId))
		mess.WriteString(fmt.Sprintf("link: %s\n\n", b.ChLink))
	}

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

func (srv *TgService) sendForceReply(chat int, mess string) error {
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
