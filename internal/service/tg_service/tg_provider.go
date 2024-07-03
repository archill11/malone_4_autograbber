package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"myapp/internal/models"
	"net/http"
	"strconv"
)

func (srv *TgService) GetUpdates(offset, timeout int, token string) ([]models.Update, error) {
	json_data, err := json.Marshal(map[string]any{
		"offset":  offset,
		"timeout": timeout,
	})
	if err != nil {
		return []models.Update{}, fmt.Errorf("GetUpdates Marshal err: %v", err)
	}
	resp, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgLocEndp, token, "getUpdates"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return []models.Update{}, fmt.Errorf("GetUpdates Post err: %v", err)
	}
	defer resp.Body.Close()
	var cAny models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&cAny); err != nil {
		return cAny.Result, fmt.Errorf("GetUpdates Decode err: %v", err)
	}
	if cAny.ErrorCode != 0 {
		return cAny.Result, fmt.Errorf("GetUpdates errResp: %+v", cAny.BotErrResp)
	}
	return cAny.Result, nil
}

func (srv *TgService) GetMe(token string) (models.ApiBotResp, error) {
	resp, err := http.Get(fmt.Sprintf(srv.Cfg.TgEndp, token, "getMe"))
	if err != nil {
		return models.ApiBotResp{}, fmt.Errorf("GetMe Get err: %v", err)
	}
	defer resp.Body.Close()
	var cAny models.ApiBotResp
	if err := json.NewDecoder(resp.Body).Decode(&cAny); err != nil {
		return models.ApiBotResp{}, err
	}
	if cAny.ErrorCode != 0 {
		return cAny, fmt.Errorf("GetMe errResp: %+v", cAny)
	}
	return cAny, nil
}

func (srv *TgService) GetChat(chatId int, token string) (models.GetChatResp, error) {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
	})
	if err != nil {
		return models.GetChatResp{}, err
	}
	resp, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgEndp, token, "getChat"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return models.GetChatResp{}, err
	}
	defer resp.Body.Close()
	var cAny models.GetChatResp
	if err := json.NewDecoder(resp.Body).Decode(&cAny); err != nil {
		return models.GetChatResp{}, err
	}
	if cAny.ErrorCode != 0 {
		return cAny, fmt.Errorf("GetChat errResp: %+v", cAny)
	}
	return cAny, nil
}

func (srv *TgService) GetFile(fileId string) (models.GetFileResp, error) {
	resp, err := http.Get(
		fmt.Sprintf(srv.Cfg.TgLocEndp, srv.Cfg.Token, fmt.Sprintf("getFile?file_id=%s", fileId)),
	)
	if err != nil {
		return models.GetFileResp{}, fmt.Errorf("GetFile Get file_id-%s err: %v", fileId, err)
	}
	defer resp.Body.Close()
	var cAny models.GetFileResp
	if err := json.NewDecoder(resp.Body).Decode(&cAny); err != nil {
		return models.GetFileResp{}, fmt.Errorf("GetFile Decode err: %v", err)
	}
	if cAny.ErrorCode != 0 {
		err = fmt.Errorf("GetFile errResp: %+v", cAny)
		if cAny.Description == "Bad Request: invalid file_id" {
			err = fmt.Errorf("%v\n\n\nТГ СЕРВЕР ИЗМЕНЕН НА ОБЫЧНЫЙ api.telegram.org (не локальный %s)", err, srv.Cfg.TgLocUrl)
			srv.Cfg.TgUrl = "https://api.telegram.org"
			srv.Cfg.TgEndp = "https://api.telegram.org/bot%s/%s"
			srv.Cfg.TgLocUrl = "https://api.telegram.org"
			srv.Cfg.TgLocEndp = "https://api.telegram.org/bot%s/%s"
		}
		return cAny, err
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
	err = srv.sendData(json_data, "sendMessage")
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) SendMessage(chat int, mess string) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chat),
		"text":    mess,
		"parse_mode": "HTML",
		"link_preview_options": `{"is_disabled": true}`,
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

func (srv *TgService) SendMessageByToken(chat int, mess string, token string) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chat),
		"text":    mess,
		"parse_mode": "HTML",
		"link_preview_options": `{"is_disabled": true}`,
	})
	if err != nil {
		return err
	}
	err = srv.sendData_v2(json_data, token, "sendMessage")
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) SendMediaGroup(json_data []byte, token string) (models.SendMediaGroupResp, error) {
	resp, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgLocEndp, token, "sendMediaGroup"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return models.SendMediaGroupResp{}, fmt.Errorf("SendMediaGroup Post %v", err)
	}
	defer resp.Body.Close()
	var sendMGResp models.SendMediaGroupResp
	if err := json.NewDecoder(resp.Body).Decode(&sendMGResp); err != nil {
		return models.SendMediaGroupResp{}, fmt.Errorf("SendMediaGroup Decode err: %v", err)
	}
	if sendMGResp.ErrorCode != 0 {
		return sendMGResp, fmt.Errorf("SendMediaGroup BotErrResp: %v", sendMGResp.BotErrResp)
	}
	return sendMGResp, nil
}

func (srv *TgService) SendVideoNote(body io.Reader, contentType string, token string) (models.SendMediaResp, error) {
	resp, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgLocEndp, token, "sendVideoNote"),
		contentType,
		body,
	)
	if err != nil {
		return models.SendMediaResp{}, fmt.Errorf("SendVideoNote Post %v", err)
	}
	defer resp.Body.Close()
	var sendMGResp models.SendMediaResp
	if err := json.NewDecoder(resp.Body).Decode(&sendMGResp); err != nil {
		return models.SendMediaResp{}, fmt.Errorf("SendVideoNote Decode err: %v", err)
	}
	if sendMGResp.ErrorCode != 0 {
		return sendMGResp, fmt.Errorf("SendVideoNote BotErrResp: %v", sendMGResp.BotErrResp)
	}
	return sendMGResp, nil
}

func (srv *TgService) DeleteMessage(chat, messId int, token string) error {
	srv.l.Info(fmt.Sprintf("DeleteMessage chat_id: %d, message_id: %d, token: %s", chat, messId, token))
	json_data, err := json.Marshal(map[string]any{
		"chat_id":    strconv.Itoa(chat),
		"message_id": strconv.Itoa(messId),
	})
	if err != nil {
		return err
	}
	err = srv.sendData_v2(json_data, token, "deleteMessage")
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) EditMessageText(json_data []byte, botToken string) error {
	err := srv.sendData_v2(json_data, botToken, "editMessageText")
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) EditMessageCaption(json_data []byte, botToken string) error {
	err := srv.sendData_v2(json_data, botToken, "editMessageCaption")
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) sendData(json_data []byte, method string) error {
	resp, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgEndp, srv.Cfg.Token, method),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return fmt.Errorf("sendData Post err: %v", err)
	}
	defer resp.Body.Close()
	var cAny models.BotErrResp
	if err := json.NewDecoder(resp.Body).Decode(&cAny); err != nil {
		return fmt.Errorf("sendData Decode err: %v", err)
	}
	if cAny.ErrorCode != 0 {
		return fmt.Errorf("sendData ErrResp: %+v", cAny)
	}
	return nil
}

func (srv *TgService) sendData_v2(json_data []byte, botToken, method string) error {
	resp, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgEndp, botToken, method),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return fmt.Errorf("sendData_v2 Post err: %v", err)
	}
	defer resp.Body.Close()
	var cAny models.BotErrResp
	if err := json.NewDecoder(resp.Body).Decode(&cAny); err != nil {
		return fmt.Errorf("sendData_v2 Decode err: %v", err)
	}
	if cAny.ErrorCode != 0 {
		return fmt.Errorf("sendData_v2 ErrResp: %+v", cAny)
	}
	return nil
}
