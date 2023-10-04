package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myapp/internal/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (srv *TgService) UB_add_channel(link string) (models.UB_add_channel_resp, error) {
	srv.l.Info(fmt.Sprintf("send request UB_add_channel /add_new_channel link-%s", link))

	json_data, err := json.Marshal(map[string]any{
		"url": link,
	})
	if err != nil {
		return models.UB_add_channel_resp{}, fmt.Errorf("UB_add_channel Marshal err: %v", err)
	}
	resp, err := http.Post(
		fmt.Sprintf("%s/%s", srv.Cfg.UserbotHost, "add_new_channel"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return models.UB_add_channel_resp{}, fmt.Errorf("UB_add_channel http.Post err: %v", err)
	}
	defer resp.Body.Close()
	var chResp models.UB_add_channel_resp
	if err := json.NewDecoder(resp.Body).Decode(&chResp); err != nil {
		return chResp, fmt.Errorf("UB_add_channel Decode Body err: %v", err)
	}
	if chResp.Status == "" {
		return chResp, fmt.Errorf("UB_add_channel err: пустой запрос от юзербота %+v", chResp)
	}
	return chResp, nil
}

//_____________________________________________________________________FLOOD_______________________________________________________________________________________

func (srv *TgService) UB_add_channel_flood(link string, fromId int) (models.UB_add_channel_resp, error) {
	addChannelResp, err := srv.UB_add_channel(link)
	if err != nil {
		err := fmt.Errorf("UB_add_channel_flood UB_add_channel err: %v", err)
		srv.l.Error(err.Error())
		srv.SendMessage(fromId, err.Error())
		return addChannelResp, err
	}
	for {
		floodWaitSec, err := srv.ParseFloodWait(addChannelResp.Error)
		if err != nil {
			errMess := fmt.Errorf("UB_add_channel_flood ParseFoodWaitV2 err: %+v", err)
			srv.SendMessage(fromId, errMess.Error())
			srv.l.Error(errMess.Error())
		}
		if floodWaitSec == 0 {
			break
		}
		logMess := fmt.Sprintf("🥶 UB_add_channel_flood исчерпан лимит запросов, жду %d секунд и повторяю запрос UB_add_channel(%s)", floodWaitSec, link)
		srv.SendMessage(fromId, logMess)
		srv.l.Info(logMess)
		time.Sleep(time.Second * time.Duration(floodWaitSec))

		addChannelResp, err = srv.UB_add_channel(link)
		if err != nil {
			err := fmt.Errorf("UB_add_channel_flood UB_add_channel err: %v", err)
			srv.SendMessage(fromId, err.Error())
			srv.l.Error(err.Error())
		}
	}
	return addChannelResp, nil
}

func (srv *TgService) ParseFloodWait(errMess string) (int, error) {
	if strings.HasPrefix(errMess, "FLOOD_WAIT_") {
		errArr := strings.Split(errMess, "FLOOD_WAIT_")
		if len(errArr) != 2 {
			return 0, fmt.Errorf("ParseFoodWait err: len(errArr) != 2, ErrorMessage-%s", errMess)
		}
		waitSecondsStr := errArr[1]
		waitSeconds, err := strconv.Atoi(waitSecondsStr)
		if err != nil {
			return 0, fmt.Errorf("ParseFoodWait Atoi err: %+v, waitSecondsStr-%s, ErrorMessage-%s", err, waitSecondsStr, errMess)
		}
		return waitSeconds, nil
	}
	return 0 , nil
}