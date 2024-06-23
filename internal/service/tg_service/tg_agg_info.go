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
		user, _ := srv.db.GetUserById(b.UserCreator)
		mess.WriteString(fmt.Sprintf("	user: %s\n", fmt.Sprintf("%d | %s", b.UserCreator, srv.AddAt(user.Username))))
		mess.WriteString(fmt.Sprintf("	личка: %s\n", b.Lichka))

		if i%30 == 0 && i > 0 {
			err = srv.SendMessage(chatId, mess.String())
			if err != nil {
				return err
			}
			mess.Reset()
		}
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
	err = srv.sendData(json_data, "sendMessage")
	if err != nil {
		return fmt.Errorf("showBotsAndChannels sendData err: %v", err)
	}
	return nil
}

func (srv *TgService) showBotsAndChannels_user(chatId int) error {
	bots, err := srv.db.GetAllBots()
	if err != nil {
		return err
	}
	var mess bytes.Buffer
	mess.WriteString("Ваши боты:\n")
	for i, b := range bots {
		if b.UserCreator != chatId {
			continue
		}
		mess.WriteString(fmt.Sprintf("%d) id: %d - @%s ", i+1, b.Id, b.Username))
		if b.IsDonor == 1 {
			mess.WriteString("-Донор")
		}
		mess.WriteString(fmt.Sprintf("\n	ch_link: %s\n", b.ChLink))
		mess.WriteString(fmt.Sprintf("	личка: %s\n", b.Lichka))

		if i%30 == 0 && i > 0 {
			err = srv.SendMessage(chatId, mess.String())
			if err != nil {
				return err
			}
			mess.Reset()
		}
	}
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
		"text":    mess.String(),
		"reply_markup": `{"inline_keyboard" : [
			[{ "text": "Назад", "callback_data": "show_user_panel" }]
		]}`,
	})
	if err != nil {
		return err
	}
	err = srv.sendData(json_data, "sendMessage")
	if err != nil {
		return fmt.Errorf("showBotsAndChannels_user sendData err: %v", err)
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
		mess.WriteString(fmt.Sprintf("Количество Привязаных ботов: %d\n\n", len(bots)))

		if i%30 == 0 && i > 0 {
			err = srv.SendMessage(chatId, mess.String())
			if err != nil {
				return err
			}
			mess.Reset()
		}
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
	err = srv.sendData(json_data, "sendMessage")
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) showAllGroupLinks_user(chatId int) error {
	grs, err := srv.db.GetAllGroupLinks()
	if err != nil {
		return err
	}
	var mess bytes.Buffer
	mess.WriteString("Ваши гр-ссылки:\n")
	for i, b := range grs {
		if b.UserCreator != chatId {
			continue
		}
		mess.WriteString(fmt.Sprintf("%d) id: %d\n", i+1, b.Id))
		mess.WriteString(fmt.Sprintf("Название: %s\n", b.Title))
		mess.WriteString(fmt.Sprintf("Ссылка: %s\n", b.Link))
		bots, err := srv.db.GetBotsByGrouLinkId(b.Id)
		if err != nil {
			return err
		}
		mess.WriteString(fmt.Sprintf("Количество Привязаных ботов: %d\n\n", len(bots)))

		if i%30 == 0 && i > 0 {
			err = srv.SendMessage(chatId, mess.String())
			if err != nil {
				return err
			}
			mess.Reset()
		}
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
	err = srv.sendData(json_data, "sendMessage")
	if err != nil {
		return err
	}
	return nil
}
