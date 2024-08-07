package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

func (srv *TgService) showAdminPanel(chatId int) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
		"text":    "Привет, я бот Донор",
		"reply_markup": `{"inline_keyboard" : [
			[{ "text": "Привязанные боты и каналы", "callback_data": "show_bots_and_channels" }],
			[{ "text": "➕ Добавить бота", "callback_data": "create_vampere_bot" }],
			[{ "text": "🗑 Удалить бота", "callback_data": "delete_vampere_bot" }],
			[{ "text": "➕ Добавить канал боту", "callback_data": "add_ch_to_bot" }],
			[{ "text": "➕ Добавить группу-ссылку", "callback_data": "create_group_link" }],
			[{ "text": "🗑 Удалить группу-ссылку", "callback_data": "delete_group_link" }],
			[{ "text": "🖌 Редактировать группу-ссылку", "callback_data": "update_group_link" }],
			[{ "text": "🖌 Поменять группу-ссылку у бота", "callback_data": "edit_bot_group_link" }],
			[{ "text": "Все группы-ссылки", "callback_data": "show_all_group_links" }],
			[{ "text": "🖌 Поменять личку у бота", "callback_data": "edit_bot_lichka" }],
			[{ "text": "🖌 Поменять личку по группе-ссылке", "callback_data": "edit_bot_lichka_by_group_link" }],
			 [{ "text": "🖌 Поменять личку везде", "callback_data": "edit_bot_lichka_all" }],
			[{ "text": "➕ Добавить Админа", "callback_data": "add_admin_btn" }],
			[{ "text": "➕ Добавить Юзера", "callback_data": "add_user_btn" }],
			[{ "text": "🗑 Удалить Юзера", "callback_data": "del_user_btn" }],
			[{ "text": "🗑 Удалить пост во всех каналах", "callback_data": "del_post_in_chs_bots" }],
			[{ "text": "🗑 Удалить потеряных ботов", "callback_data": "del_lost_bots" }],
			[{ "text": "Restart app", "callback_data": "restart_app" }]
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

func (srv *TgService) showUserPanel(chatId int) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
		"text":    "Привет, я бот Донор",
		"reply_markup": `{"inline_keyboard" : [
			[{ "text": "Привязанные боты и каналы", "callback_data": "show_bots_and_channels_user" }],
			[{ "text": "➕ Добавить бота", "callback_data": "create_vampere_bot" }],
			[{ "text": "🗑 Удалить бота", "callback_data": "delete_vampere_bot" }],
			[{ "text": "➕ Добавить канал боту", "callback_data": "add_ch_to_bot" }],
			[{ "text": "➕ Добавить группу-ссылку", "callback_data": "create_group_link" }],
			[{ "text": "🗑 Удалить группу-ссылку", "callback_data": "delete_group_link" }],
			[{ "text": "🖌 Редактировать группу-ссылку", "callback_data": "update_group_link" }],
			[{ "text": "🖌 Поменять группу-ссылку у бота", "callback_data": "edit_bot_group_link" }],
			[{ "text": "Все группы-ссылки", "callback_data": "show_all_group_links_user" }],
			[{ "text": "🖌 Поменять личку у бота", "callback_data": "edit_bot_lichka" }]
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

func (srv *TgService) showAdminPanelRoles(chatId int) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
		"text":    "Привет, я бот Донор",
		"reply_markup": `{"inline_keyboard" : [
			[{ "text": "Заменить домен", "callback_data": "change_domen_btn" }],
			[{ "text": "➕ Добавить админа", "callback_data": "add_admin_btn" }],
			[{ "text": "🗑 Удалить админа", "callback_data": "del_admin_btn" }]
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

func (srv *TgService) showCfgPanel(chatId int) error {
	var rm bytes.Buffer
	rm.WriteString(`{"inline_keyboard" : [`)
	cfgVal, _ := srv.db.GetCfgValById("auto-acc-media-gr")
	if cfgVal.Val == "1" {
		rm.WriteString(`[{ "text": "выкл авто подтвержение", "callback_data": "change_auto-acc-media-gr_to_0_btn" }],`)
	} else {
		rm.WriteString(`[{ "text": "вкл авто подтвержение", "callback_data": "change_auto-acc-media-gr_to_1_btn" }],`)
	}
	rm.WriteString(`[{ "text": "________", "callback_data": "_____" }]`)
	rm.WriteString(`]}`)

	var mess bytes.Buffer
	mess.WriteString(fmt.Sprintf("авто подтвержение: %s\n", cfgVal.Val))

	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
		"text":    mess.String(),
		"reply_markup": rm.String(),
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
