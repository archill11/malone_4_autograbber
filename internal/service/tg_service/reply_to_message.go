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
)

func (srv *TgService) RM_obtain_vampire_bot_token(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service::tg::rm::", rm.Text, replyMes)

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
	resp, err := http.Get(fmt.Sprintf(srv.TgEndp, bot.Token, fmt.Sprintf("setWebhook?url=%s/api/v1/vampire/update", srv.HostUrl))) // set Webhook
	if err != nil {
		srv.l.Err("Error set Webhook::", err)
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
		srv.l.Err(err, tgResp.Description)
		srv.ShowMessClient(chatId, u.ERR_MSG)
	}
	srv.l.Info("set Webhook")
	err = srv.ShowMessClient(chatId, u.SUCCESS_ADDED_BOT)

	return err
}

func (srv *TgService) RM_delete_bot(m models.Update) error {
	rm := m.Message.ReplyToMessage
	replyMes := m.Message.Text
	chatId := m.Message.From.Id
	srv.l.Info("tg_service::tg::rm::", rm.Text, replyMes)

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
	_, err = http.Get(fmt.Sprintf(srv.TgEndp, bot.Token, "setWebhook?url=")) // delete Webhook
	if err != nil {
		srv.l.Err("Error delete Webhook::", err)
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
	srv.l.Info("tg_service::tg::rm::", rm.Text, replyMes)

	usr, err := srv.As.GetUserByUsername(strings.TrimSpace(replyMes))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			srv.ShowMessClient(chatId, "я не знаю такого юзера , пусть напишет мне /start")
			return err
		}
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.As.EditAdmin(usr.Username, 1)
	if err != nil {
		srv.ShowMessClient(chatId, u.ERR_MSG)
		return err
	}
	err = srv.ShowMessClient(chatId, "Админ добавлен")
	return err
}
