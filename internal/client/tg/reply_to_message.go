package tg

import (
	"myapp/internal/models"
	u "myapp/internal/utils"
	"strconv"
	"strings"
)

func (srv *TgClient) HandleReplyToMessage(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	// chatId := m.Message.From.Id
	// srv.l.Info("client::tg::rm::", m.Message)
	srv.l.Info("client::tg::rm::", rm.Text, replyMes)

	if rm.Text == u.NEW_BOT_MSG {
		err := srv.Ts.RM_obtain_vampire_bot_token(m)
		return err
	}

	if rm.Text == u.DELETE_BOT_MSG {
		err := srv.Ts.RM_delete_bot(m)
		return err
	}

	if rm.Text == u.NEW_ADMIN_MSG {
		err := srv.Ts.RM_add_admin(m)
		return err
	}

	if rm.Text == u.NEW_GROUP_LINK_MSG {
		err := srv.Ts.RM_add_group_link(m)
		return err
	}

	if rm.Text == u.DELETE_GROUP_LINK_MSG {
		err := srv.Ts.RM_delete_group_link(m)
		return err
	}

	if strings.HasPrefix(rm.Text, "укажите номер группы-ссылки для нового бота[") {
		runes := []rune(rm.Text)
		runesStr := string(runes[len([]rune("укажите номер группы-ссылки для нового бота[")):])
		botId, _ := strconv.Atoi(runesStr)
		err := srv.Ts.RM_update_bot_group_link(m, botId)
		return err
	}

	return nil
}
