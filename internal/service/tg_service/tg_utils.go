package tg_service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"myapp/internal/entity"
	"myapp/internal/models"
	"myapp/pkg/files"
	"myapp/pkg/mycopy"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func (srv *TgService) showAdminPanel(chatId int) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
		"text":    "Привет, я бот Донор",
		"reply_markup": `{"inline_keyboard" : [
			[{ "text": "Создать бота", "callback_data": "create_vampere_bot" }],
			[{ "text": "Удалить бота", "callback_data": "delete_vampere_bot" }]
		]}`,
	})
	if err != nil {
		return err
	}
	err = srv.sendData(json_data)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) getBotByToken(token string) (models.APIRBotresp, error) {
	resp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, token, "getMe"),
	)
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




func (srv *TgService) sendChPostAsVamp(vampBot entity.Bot, m models.Update) error {
	donor_ch_mes_id := m.ChannelPost.MessageId

	if m.ChannelPost.VideoNote.FileId != "" {
		//////////////// если кружочек видео
		err := srv.sendChPostAsVamp_VideoNote(vampBot, m)
		return err
	}else if len(m.ChannelPost.Photo) > 0 {
		//////////////// если фото
		err := srv.sendChPostAsVamp_Photo(vampBot, m)
		return err
	}else if m.ChannelPost.Video.FileId != "" {
		//////////////// если видео
		err := srv.sendChPostAsVamp_Video(vampBot, m)
		return err
	}else{
		//////////////// если просто текст
		fmt.Println("Just Message !!!!")
		futureMesJson := map[string]any{
			"chat_id": strconv.Itoa(vampBot.ChId),
		}
		if m.ChannelPost.ReplyToMessage.MessageId != 0 {
			fmt.Println("ReplyToMessage !!!!")
			// ReplToDonorChId := m.ChannelPost.ReplyToMessage.Chat.Id
			replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
			currPost, err := srv.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
			if err != nil {
				return err
			}
			futureMesJson["reply_to_message_id"] = currPost.PostId
		}
		if len(m.ChannelPost.Entities) > 0 {
			fmt.Println("Entities !!!!")
			entities := make([]models.MessageEntity, len(m.ChannelPost.Entities))
			mycopy.DeepCopy(m.ChannelPost.Entities, &entities)
			for i, v := range entities {
				urlArr := strings.Split(v.Url, "/")
				for ii, vv := range urlArr {
					if vv == "t.me" && urlArr[ii+1] == "c" {
						fmt.Printf("\nэто ссылка на канал %s и пост %s\n", urlArr[ii+2], urlArr[ii+3])
						refToDonorChPostId, err := strconv.Atoi(urlArr[ii+3])
						if err != nil {
							return err
						}
						currPost, err := srv.As.GetPostByDonorIdAndChId(refToDonorChPostId, vampBot.ChId)
						if err != nil {
							return err
						}
						if vampBot.ChId < 0 {
							urlArr[ii+2] = strconv.Itoa(-vampBot.ChId)
						} else {
							urlArr[ii+2] = strconv.Itoa(vampBot.ChId)
						}
						if urlArr[ii+2][0] == '1' && urlArr[ii+2][1] == '0' && urlArr[ii+2][2] == '0' {
							urlArr[ii+2] = urlArr[ii+2][3:]
						}
						urlArr[ii+3] = strconv.Itoa(currPost.PostId)
						entities[i].Url = strings.Join(urlArr, "/")
					}
				}
			}
			futureMesJson["entities"] = entities
		}
	
		text_message := m.ChannelPost.Text
		futureMesJson["text"] = text_message
		json_data, err := json.Marshal(futureMesJson)
		if err != nil {
			return err
		}
		sendVampPostResp, err := http.Post(
			fmt.Sprintf(srv.TgEndp, vampBot.Token, "sendMessage"),
			"application/json",
			bytes.NewBuffer(json_data),
		)
		if err != nil {
			return err
		}
		defer sendVampPostResp.Body.Close()
		var cAny struct {
			Ok     bool `json:"ok"`
			Result struct {
				MessageId int `json:"message_id"`
			} `json:"result,omitempty"`
		}
		if err := json.NewDecoder(sendVampPostResp.Body).Decode(&cAny); err != nil {
			return err
		}
		if cAny.Result.MessageId != 0 {
			err = srv.As.AddNewPost(vampBot.ChId, cAny.Result.MessageId, donor_ch_mes_id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (srv *TgService) sendChPostAsVamp_VideoNote(vampBot entity.Bot, m models.Update) error {
	fmt.Println("VideoNote !!!!")
	donor_ch_mes_id := m.ChannelPost.MessageId
	futureVideoNoteJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	if m.ChannelPost.ReplyToMessage.MessageId != 0 {
		fmt.Println("ReplyToMessage !!!!")
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
		if err != nil {
			return err
		}
		futureVideoNoteJson["reply_to_message_id"] = strconv.Itoa(currPost.PostId)
	}
	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", m.ChannelPost.VideoNote.FileId)),
	)
	if err != nil {
		return err
	}
	defer getFilePAthResp.Body.Close()
	var cAny struct {
		Ok     bool `json:"ok"`
		Result struct {
			File_id string `json:"file_id"`
			File_unique_id string `json:"file_unique_id"`
			File_path string `json:"file_path"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(getFilePAthResp.Body).Decode(&cAny); err != nil {
		return err
	}
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	fmt.Println(fileNameInServer)
	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path),
		)
		if err != nil {
			return err
		}
	}
	futureVideoNoteJson["video_note"] = fmt.Sprintf("@%s", fileNameInServer)
	cf, body, err := files.CreateForm(futureVideoNoteJson)
	if err != nil {
		return err
	}
	rrres, err := http.Post(
		fmt.Sprintf(srv.TgEndp, vampBot.Token, "sendVideoNote"),
		cf, 
		body,
	)
	if err != nil {
		return err
	}
	fmt.Sprintln()
	defer rrres.Body.Close()
	var cAny2 struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageId int `json:"message_id"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return err
	}
	if cAny2.Result.MessageId != 0 {
		err = srv.As.AddNewPost(vampBot.ChId, cAny2.Result.MessageId, donor_ch_mes_id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *TgService) sendChPostAsVamp_Photo(vampBot entity.Bot, m models.Update) error {
	if m.ChannelPost.MediaGroupId != "" {
		go srv.sendChPostAsVamp_Photo_MediaGroup(vampBot, m)
		return nil
	}
	fmt.Println("Photo !!!!")
	donor_ch_mes_id := m.ChannelPost.MessageId
	futurePhotoJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}

	if m.ChannelPost.ReplyToMessage.MessageId != 0 {
		fmt.Println("ReplyToMessage !!!!")
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
		if err != nil {
			return err
		}
		futurePhotoJson["reply_to_message_id"] = strconv.Itoa(currPost.PostId)
	}
	if m.ChannelPost.Caption != "" {
		futurePhotoJson["caption"] = m.ChannelPost.Caption
	}
	if len(m.ChannelPost.CaptionEntities) > 0 {
		fmt.Println("CaptionEntities !!!!")
			entities := make([]models.MessageEntity, len(m.ChannelPost.CaptionEntities))
			mycopy.DeepCopy(m.ChannelPost.CaptionEntities, &entities)
			for i, v := range entities {
				urlArr := strings.Split(v.Url, "/")
				for ii, vv := range urlArr {
					if vv == "t.me" && urlArr[ii+1] == "c" {
						fmt.Printf("\nэто ссылка на канал %s и пост %s\n", urlArr[ii+2], urlArr[ii+3])
						refToDonorChPostId, err := strconv.Atoi(urlArr[ii+3])
						if err != nil {
							return err
						}
						currPost, err := srv.As.GetPostByDonorIdAndChId(refToDonorChPostId, vampBot.ChId)
						if err != nil {
							return err
						}
						if vampBot.ChId < 0 {
							urlArr[ii+2] = strconv.Itoa(-vampBot.ChId)
						} else {
							urlArr[ii+2] = strconv.Itoa(vampBot.ChId)
						}
						if urlArr[ii+2][0] == '1' && urlArr[ii+2][1] == '0' && urlArr[ii+2][2] == '0' {
							urlArr[ii+2] = urlArr[ii+2][3:]
						}
						urlArr[ii+3] = strconv.Itoa(currPost.PostId)
						entities[i].Url = strings.Join(urlArr, "/")
					}
				}
			}
			j, _ := json.Marshal(entities)
			futurePhotoJson["caption_entities"] = string(j)
	}

	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId)),
	)
	if err != nil {
		return err
	}
	defer getFilePAthResp.Body.Close()
	var cAny struct {
		Ok     bool `json:"ok"`
		Result struct {
			File_id string `json:"file_id"`
			File_unique_id string `json:"file_unique_id"`
			File_path string `json:"file_path"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(getFilePAthResp.Body).Decode(&cAny); err != nil {
		return err
	}
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	fmt.Println("fileNameInServer:",fileNameInServer)
	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path),
		)
		if err != nil {
			return err
		}
	}
	futurePhotoJson["photo"] = fmt.Sprintf("@%s", fileNameInServer)
	futurePhotoJson["disable_notification"] = "true"
	cf, body, err := files.CreateForm(futurePhotoJson)
	if err != nil {
		return err
	}
	rrres, err := http.Post(
		fmt.Sprintf(srv.TgEndp, vampBot.Token, "sendPhoto"),
		cf,
		body,
	)
	if err != nil {
		return err
	}
	fmt.Sprintln()
	defer rrres.Body.Close()
	var cAny2 struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageId int `json:"message_id"`
			Chat  struct {Id int `json:"id"`} `json:"chat"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return err
	}
	if cAny2.Result.MessageId != 0 {
		err = srv.As.AddNewPost(vampBot.ChId, cAny2.Result.MessageId, donor_ch_mes_id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *TgService) sendChPostAsVamp_Photo_MediaGroup(vampBot entity.Bot, m models.Update) error {
	fmt.Println("Photo_MediaGroup !!!!")

	donor_ch_mes_id := m.ChannelPost.MessageId
	futurePhotoJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId)),
	)
	if err != nil {
		return err
	}
	defer getFilePAthResp.Body.Close()
	var cAny struct {
		Ok     bool `json:"ok"`
		Result struct {
			File_id string `json:"file_id"`
			File_unique_id string `json:"file_unique_id"`
			File_path string `json:"file_path"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(getFilePAthResp.Body).Decode(&cAny); err != nil {
		return err
	}
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	fmt.Println("fileNameInServer:",fileNameInServer)
	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path),
		)
		if err != nil {
			return err
		}
	}
	futurePhotoJson["photo"] = fmt.Sprintf("@%s", fileNameInServer)
	futurePhotoJson["disable_notification"] = "true"
	cf, body, err := files.CreateForm(futurePhotoJson)
	if err != nil {
		return err
	}
	rrres, err := http.Post(
		fmt.Sprintf(srv.TgEndp, vampBot.Token, "sendPhoto"),
		cf,
		body,
	)
	if err != nil {
		return err
	}
	defer rrres.Body.Close()
	var cAny2 struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageId int `json:"message_id"`
			Chat  struct {Id int `json:"id"`} `json:"chat"`
			Photo []models.PhotoSize `json:"photo"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return err
	}

	PhotoDelJson, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(vampBot.ChId),
		"message_id": strconv.Itoa(cAny2.Result.MessageId),
	})
	if err != nil {
		return err
	}
	_, err = http.Post(
		fmt.Sprintf(srv.TgEndp, vampBot.Token, "deleteMessage"),
		"application/json",
		bytes.NewBuffer(PhotoDelJson),
	)
	if err != nil {
		return err
	}

	srv.LMG.Mu.Lock()
	srv.LMG.MuExecuted = true
	defer func() {
		srv.LMG.Mu.Unlock()
		srv.LMG.MuExecuted = false
	}()
	newmedia := Media{
		Media_group_id: m.ChannelPost.MediaGroupId,
		Type_media: "photo",
		File_id: cAny2.Result.Photo[len(cAny2.Result.Photo)-1].FileId,
		Donor_message_id: donor_ch_mes_id,
	}
	if m.ChannelPost.Caption != "" {
		newmedia.Caption = m.ChannelPost.Caption
	}
	if m.ChannelPost.ReplyToMessage.MessageId != 0 {
		fmt.Println("ReplyToMessage !!!!")
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
		if err != nil {
			return err
		}
		newmedia.Reply_to_message_id = currPost.PostId
	}
	if len(m.ChannelPost.CaptionEntities) > 0 {
		fmt.Println("CaptionEntities !!!!")
		entities := make([]models.MessageEntity, len(m.ChannelPost.CaptionEntities))
		mycopy.DeepCopy(m.ChannelPost.CaptionEntities, &entities)
		for i, v := range entities {
			urlArr := strings.Split(v.Url, "/")
			for ii, vv := range urlArr {
				if vv == "t.me" && urlArr[ii+1] == "c" {
					fmt.Printf("\nэто ссылка на канал %s и пост %s\n", urlArr[ii+2], urlArr[ii+3])
					refToDonorChPostId, err := strconv.Atoi(urlArr[ii+3])
					if err != nil {
						return err
					}
					currPost, err := srv.As.GetPostByDonorIdAndChId(refToDonorChPostId, vampBot.ChId)
					if err != nil {
						return err
					}
					if vampBot.ChId < 0 {
						urlArr[ii+2] = strconv.Itoa(-vampBot.ChId)
					} else {
						urlArr[ii+2] = strconv.Itoa(vampBot.ChId)
					}
					if urlArr[ii+2][0] == '1' && urlArr[ii+2][1] == '0' && urlArr[ii+2][2] == '0' {
						urlArr[ii+2] = urlArr[ii+2][3:]
					}
					urlArr[ii+3] = strconv.Itoa(currPost.PostId)
					entities[i].Url = strings.Join(urlArr, "/")
				}
			}
		}
		newmedia.Caption_entities = entities
	}
	srv.LMG.MediaGroups[fmt.Sprintf("%s:%s", m.ChannelPost.MediaGroupId, strconv.Itoa(vampBot.ChId))] = append(srv.LMG.MediaGroups[fmt.Sprintf("%s:%s", m.ChannelPost.MediaGroupId, strconv.Itoa(vampBot.ChId))], newmedia)
	srv.LMG.Mu.Unlock()
	srv.LMG.MuExecuted = false
	time.Sleep(time.Second*5)
	srv.LMG.Mu.Lock()
	srv.LMG.MuExecuted = true
	medias, ok := srv.LMG.MediaGroups[fmt.Sprintf("%s:%s", m.ChannelPost.MediaGroupId, strconv.Itoa(vampBot.ChId))]
	if !ok {
		return nil 
	}
	if len(medias) < 2 {
		srv.LMG.Mu.Unlock()
		srv.LMG.MuExecuted = false
		time.Sleep(time.Second*5)
	}
	fmt.Println("len(medias):::", len(medias))
	srv.LMG.Mu.TryLock()
	srv.LMG.MuExecuted = true
	medias2, ok := srv.LMG.MediaGroups[fmt.Sprintf("%s:%s", m.ChannelPost.MediaGroupId, strconv.Itoa(vampBot.ChId))]
	if !ok {
		return nil 
	}
	s2 := make([]Media, len(medias2))
	copy(s2, medias2)
	delete(srv.LMG.MediaGroups, fmt.Sprintf("%s:%s", m.ChannelPost.MediaGroupId, strconv.Itoa(vampBot.ChId)))
	srv.LMG.Mu.Unlock()
	srv.LMG.MuExecuted = false

	arrsik := make([]models.InputMediaPhoto, 0)
	for _, med := range s2 {
		nwmd := models.InputMediaPhoto{
			Type: med.Type_media,
			Media: med.File_id,
			Caption: med.Caption,
			CaptionEntities: med.Caption_entities,
		}
		arrsik = append(arrsik, nwmd)
	}
	ttttt := map[string]any{
		"chat_id": strconv.Itoa(vampBot.ChId),
		"media": arrsik,
	}
	if s2[0].Reply_to_message_id != 0 {
		ttttt["reply_to_message_id"] = s2[0].Reply_to_message_id
	}
	MediaJson, err := json.Marshal(ttttt)
	if err != nil {
		return err
	}
	rrresfyhfy, err := http.Post(
		fmt.Sprintf(srv.TgEndp, vampBot.Token, "sendMediaGroup"),
		"application/json",
		bytes.NewBuffer(MediaJson),
	)
	if err != nil {
		return err
	}
	defer rrresfyhfy.Body.Close()
	var cAny223 struct {
		Ok     bool `json:"ok"`
		Result []struct {
			MessageId int `json:"message_id"`
			Chat  struct {Id int `json:"id"`} `json:"chat"`
			Photo []models.PhotoSize `json:"photo"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrresfyhfy.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return err
	}
	for _, v := range cAny223.Result {
		if v.MessageId != 0 {
			err = srv.As.AddNewPost(vampBot.ChId, v.MessageId, donor_ch_mes_id)
			if err != nil {
				return err
			}
		}
	}









	// if m.ChannelPost.ReplyToMessage.MessageId != 0 {
	// 	fmt.Println("ReplyToMessage !!!!")
	// }
	// if len(m.ChannelPost.CaptionEntities) > 0 {
	// 	fmt.Println("CaptionEntities !!!!")
	// }
	return nil
}

func (srv *TgService) sendChPostAsVamp_Video(vampBot entity.Bot, m models.Update) error {
	if m.ChannelPost.MediaGroupId != "" {
		go srv.sendChPostAsVamp_Video_MediaGroup(vampBot, m)
		return nil
	}
	fmt.Println("Video !!!!")
	donor_ch_mes_id := m.ChannelPost.MessageId
	futureVideoJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	if m.ChannelPost.ReplyToMessage.MessageId != 0 {
		fmt.Println("ReplyToMessage !!!!")
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
		if err != nil {
			return err
		}
		futureVideoJson["reply_to_message_id"] = strconv.Itoa(currPost.PostId)
	}
	if m.ChannelPost.Caption != "" {
		futureVideoJson["caption"] = m.ChannelPost.Caption
	}
	if len(m.ChannelPost.CaptionEntities) > 0 {
		fmt.Println("CaptionEntities !!!!")
		entities := make([]models.MessageEntity, len(m.ChannelPost.CaptionEntities))
		mycopy.DeepCopy(m.ChannelPost.CaptionEntities, &entities)
		for i, v := range entities {
			urlArr := strings.Split(v.Url, "/")
			for ii, vv := range urlArr {
				if vv == "t.me" && urlArr[ii+1] == "c" {
					fmt.Printf("\nэто ссылка на канал %s и пост %s\n", urlArr[ii+2], urlArr[ii+3])
					refToDonorChPostId, err := strconv.Atoi(urlArr[ii+3])
					if err != nil {
						return err
					}
					currPost, err := srv.As.GetPostByDonorIdAndChId(refToDonorChPostId, vampBot.ChId)
					if err != nil {
						return err
					}
					if vampBot.ChId < 0 {
						urlArr[ii+2] = strconv.Itoa(-vampBot.ChId)
					} else {
						urlArr[ii+2] = strconv.Itoa(vampBot.ChId)
					}
					if urlArr[ii+2][0] == '1' && urlArr[ii+2][1] == '0' && urlArr[ii+2][2] == '0' {
						urlArr[ii+2] = urlArr[ii+2][3:]
					}
					urlArr[ii+3] = strconv.Itoa(currPost.PostId)
					entities[i].Url = strings.Join(urlArr, "/")
				}
			}
		}
		j, _ := json.Marshal(entities)
		futureVideoJson["caption_entities"] = string(j)
	}
	
	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", m.ChannelPost.Video.FileId)),
	)
	if err != nil {
		return err
	}
	defer getFilePAthResp.Body.Close()
	var cAny struct {
		Ok     bool `json:"ok"`
		Result struct {
			File_id string `json:"file_id"`
			File_unique_id string `json:"file_unique_id"`
			File_path string `json:"file_path"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(getFilePAthResp.Body).Decode(&cAny); err != nil {
		return err
	}
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	fmt.Println("fileNameInServer:",fileNameInServer)
	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path),
		)
		if err != nil {
			return err
		}
	}
	futureVideoJson["video"] = fmt.Sprintf("@%s", fileNameInServer)
	cf, body, err := files.CreateForm(futureVideoJson)
	if err != nil {
		return err
	}
	rrres, err := http.Post(
		fmt.Sprintf(srv.TgEndp, vampBot.Token, "sendVideo"),
		cf, 
		body,
	)
	if err != nil {
		return err
	}
	fmt.Sprintln()
	defer rrres.Body.Close()
	var cAny2 struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageId int `json:"message_id"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return err
	}
	if cAny2.Result.MessageId != 0 {
		err = srv.As.AddNewPost(vampBot.ChId, cAny2.Result.MessageId, donor_ch_mes_id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *TgService) sendChPostAsVamp_Video_MediaGroup(vampBot entity.Bot, m models.Update) error {
	fmt.Println("Video_MediaGroup !!!!")
	if m.ChannelPost.ReplyToMessage.MessageId != 0 {
		fmt.Println("ReplyToMessage !!!!")
	}
	if len(m.ChannelPost.CaptionEntities) > 0 {
		fmt.Println("CaptionEntities !!!!")
	}
	return nil
}















func (srv *TgService) getChatByCurrBot(chatId int, token string) (models.GetChatResult, error) {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chatId),
	})
	if err != nil {
		return models.GetChatResult{}, err
	}
	resp, err := http.Post(
		fmt.Sprintf(srv.TgEndp, token, "getChat"),
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

func (srv *TgService) sendForceReply(chat int, mess string) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id":      strconv.Itoa(chat),
		"text":         mess,
		"reply_markup": `{"force_reply": true}`,
	})
	if err != nil {
		return err
	}
	err = srv.sendData(json_data)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) ShowMessClient(chat int, mess string) error {
	json_data, err := json.Marshal(map[string]any{
		"chat_id": strconv.Itoa(chat),
		"text":    mess,
	})
	if err != nil {
		return err
	}
	err = srv.sendData(json_data)
	if err != nil {
		return err
	}
	return nil
}

func (srv *TgService) sendData(json_data []byte) error {
	_, err := http.Post(
		fmt.Sprintf(srv.TgEndp, srv.Token, "sendMessage"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return err
	}
	return nil
}

