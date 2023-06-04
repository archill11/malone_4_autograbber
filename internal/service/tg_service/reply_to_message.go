package tg_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"myapp/internal/entity"
	"myapp/internal/models"
	"myapp/internal/repository"
	u "myapp/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

func (srv *TgService) RM_obtain_vampire_bot_token(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_obtain_vampire_bot_token", zap.Any("rm.Text", rm.Text), zap.Any("replyMes",replyMes))

	tgobotResp, err := srv.getBotByToken(strings.TrimSpace(replyMes))
	if err != nil {
		return err
	}
	res := tgobotResp.Result
	bot := entity.NewBot(res.Id, res.UserName, res.FirstName, strings.TrimSpace(replyMes), 0)
	err = srv.As.AddNewBot(bot.Id, bot.Username, bot.Firstname, bot.Token, bot.IsDonor)
	if err != nil {
		return err
	}
	tgResp := struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}{}
	resp, err := http.Get(fmt.Sprintf(
		srv.TgEndp, bot.Token, fmt.Sprintf("setWebhook?url=%s/api/v1/vampire/update", srv.HostUrl)), // set Webhook
	)
	if err != nil {
		srv.l.Error("RM_obtain_vampire_bot_token: set Webhook::", zap.Error(err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &tgResp)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	if !tgResp.Ok {
		srv.l.Error("RM_obtain_vampire_bot_token: !tgResp.Ok", zap.Error(err), zap.Any("tgResp.Description", tgResp.Description))
		srv.ShowMessClient(chatId, u.ERR_MSG)
	}
	srv.l.Info("RM_obtain_vampire_bot_token: set Webhook", zap.Any("Webhook url", fmt.Sprintf(srv.TgEndp, bot.Token, fmt.Sprintf("setWebhook?url=%s/api/v1/vampire/update", srv.HostUrl))))
	srv.ShowMessClient(chatId, u.SUCCESS_ADDED_BOT)

	grl, _ := srv.As.GetAllGroupLinks()
	if len(grl) == 0 {
		return nil
	}
	err = srv.SendForceReply(chatId, fmt.Sprintf(u.GROUP_LINK_FOR_BOT_MSG, bot.Id))

	return err
}

func (srv *TgService) RM_delete_bot(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_delete_bot", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	id, err := strconv.Atoi(strings.TrimSpace(replyMes))
	if err != nil {
		srv.ShowMessClient(chatId, "неправильный формат id !")
		return err
	}
	bot, err := srv.As.GetBotInfoById(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			srv.ShowMessClient(chatId, "я не знаю такого бота !")
			return err
		}
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	if bot.IsDonor == 1 {
		srv.ShowMessClient(chatId, "главного бота нельзя удалить")
		return nil
	}
	_, err = http.Get(fmt.Sprintf(srv.TgEndp, bot.Token, "setWebhook?url=")) // delete Webhook
	if err != nil {
		srv.l.Error("RM_delete_bot: delete Webhook", zap.Error(err))
	}
	err = srv.As.DeleteBot(id)
	if err != nil {
		return err
	}
	err = srv.ShowMessClient(chatId, u.SUCCESS_DELETE_BOT)

	return err
}

func (srv *TgService) RM_add_admin(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_add_admin", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	usr, err := srv.As.GetUserByUsername(strings.TrimSpace(replyMes))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			srv.ShowMessClient(chatId, "я не знаю такого юзера , пусть напишет мне /start")
			return err
		}
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.As.GetUserByUsername(%s) : %v", strings.TrimSpace(replyMes), err)
	}
	err = srv.As.EditAdmin(usr.Username, 1)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.As.EditAdmin(%s, 1) : %v", usr.Username, err)
	}
	err = srv.ShowMessClient(chatId, "Админ добавлен")
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
		}
	}

	// link := entity.NewGroupLink(groupLinkTitle, groupLinkLink)

	err := srv.As.AddNewGroupLink(groupLinkTitle, groupLinkLink)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return fmt.Errorf("RM_add_admin: srv.As.AddNewGroupLink(%s, %s) : %v", groupLinkTitle, groupLinkLink, err)
	}
	err = srv.ShowMessClient(chatId, "группа-ссылка добавлен")
	return err
}

func (srv *TgService) RM_delete_group_link(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_delete_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))

	replyMes = strings.TrimSpace(replyMes)
	grId, err := strconv.Atoi(replyMes)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.As.DeleteGroupLink(grId)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.As.EditBotGroupLinkIdToNull(grId)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.ShowMessClient(chatId, "группа-ссылка удалена")
	return err
}

func (srv *TgService) RM_update_bot_group_link(m models.Update, botId int) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_update_bot_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))
	replyMes = strings.TrimSpace(replyMes)

	grId, err := strconv.Atoi(replyMes)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.As.EditBotGroupLinkId(grId, botId)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.ShowMessClient(chatId, fmt.Sprintf("группа-ссылка %d привязанна к боту %d", grId, botId))
	return err
}

func (srv *TgService) RM_update_group_link(m models.Update, refId int) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service: RM_update_group_link", zap.Any("rm.Text", rm.Text), zap.Any("replyMes", replyMes))
	replyMes = strings.TrimSpace(replyMes)

	err := srv.As.UpdateGroupLink(refId, replyMes)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.ShowMessClient(chatId, "группа-ссылка обновлена")
	return err
}
