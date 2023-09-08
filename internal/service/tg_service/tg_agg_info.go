package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

func (srv *TgService) showBotsAndChannels(chatId int) error {
	bots, err := srv.db.GetAllBots()
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
	grs, err := srv.db.GetAllGroupLinks()
	if err != nil {
		return err
	}
	var mess bytes.Buffer
	for i, b := range grs {
		mess.WriteString(fmt.Sprintf("%d) id: %d\n", i+1, b.Id))
		mess.WriteString(fmt.Sprintf("Название: %s\n", b.Title))
		mess.WriteString(fmt.Sprintf("Ссылка: %s\n", b.Link))
		bots, err := srv.db.GetBotsByGrouLinkId(b.Id)
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
