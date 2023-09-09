package tg_service

import (
	"encoding/json"
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
	err = srv.sendData(json_data, "sendMessage")
	if err != nil {
		return err
	}
	return nil
}
