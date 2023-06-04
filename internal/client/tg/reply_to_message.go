package tg

import (
	"fmt"
	"myapp/internal/models"
	u "myapp/internal/utils"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

func (srv *TgClient) HandleReplyToMessage(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	// chatId := m.Message.From.Id
	// srv.l.Info("client::tg::rm::", m.Message)
	srv.l.Info("tgClient: HandleReplyToMessage", zap.Any("rm.Tex", rm.Text), zap.Any("replyMes", replyMes))

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

	if rm.Text == u.UPDATE_GROUP_LINK_MSG {
		chatId := m.Message.From.Id
		replyMes := m.Message.Text
		replyMes = strings.TrimSpace(replyMes)
	
		grId, err := strconv.Atoi(replyMes)
		if err != nil {
			srv.Ts.ShowMessClient(chatId, u.ERR_MSG)
			return err
		}
		err = srv.Ts.SendForceReply(chatId, fmt.Sprintf(u.GROUP_LINK_UPDATE_MSG, grId))
		return err
	}

	if strings.HasPrefix(rm.Text, "укажите номер группы-ссылки для нового бота[") {
		runes := []rune(rm.Text)
		runesStr := string(runes[len([]rune("укажите номер группы-ссылки для нового бота[")):])
		botId, _ := strconv.Atoi(runesStr)
		err := srv.Ts.RM_update_bot_group_link(m, botId)
		return err
	}

	if strings.HasPrefix(rm.Text, "укажите новую ссылку для ref [") {
		runes := []rune(rm.Text)
		runesStr := string(runes[len([]rune("укажите новую ссылку для ref [")):])
		refId, _ := strconv.Atoi(runesStr)
		err := srv.Ts.RM_update_group_link(m, refId)
		return err
	}

	return nil
}
