package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myapp/internal/models"
	"net/http"
	"strconv"
)

func (srv *TgService) getBotByToken(token string) (models.APIRBotresp, error) {
	resp, err := http.Get(fmt.Sprintf(srv.Cfg.TgEndp, token, "getMe"))
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

func (srv *TgService) getChatByCurrBot(chatId int, token string) (models.GetChatResult, error) {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
	})
	if err != nil {
		return models.GetChatResult{}, err
	}
	resp, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgEndp, token, "getChat"),
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

func (srv *TgService) GetFile(fileId string) (models.GetFileResp, error) {
	resp, err := http.Get(
		fmt.Sprintf(srv.Cfg.TgEndp, srv.Cfg.Token, fmt.Sprintf("getFile?file_id=%s", fileId)),
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
		return models.GetFileResp{}, fmt.Errorf("GetFile errResp: %+v", cAny)
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
		"disable_web_page_preview": true,
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

func (srv *TgService) DeleteMessage(chat, messId int, botToken string) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id":    strconv.Itoa(chat),
		"message_id": strconv.Itoa(messId),
	})
	if err != nil {
		return err
	}
	err = srv.sendData_v2(json_data, botToken, "deleteMessage")
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
