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

	"go.uber.org/zap"
)

func (srv *TgService) Donor_HandleChannelPost(m models.Update) error {
	chatId := m.ChannelPost.Chat.Id
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	srv.l.Info("tgClient: Donor_HandleChannelPost", zap.Any("models.Update", m))

	err := srv.Donor_addChannelPost(m)
	if err != nil {
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG + err.Error())
			srv.SendMessage(chatId, err.Error())
		}
		return err
	}
	return nil
}

func (srv *TgService) Donor_addChannelPost(m models.Update) error {
	message_id := m.ChannelPost.MessageId
	channel_id := m.ChannelPost.Chat.Id

	// Проверка что пост есть уже в базе нужна для того что бы телега не отрпавляла
	// кучу запросов повторно , тк ответ долгий из за рассылки
	post, err := srv.db.GetPostByDonorIdAndChId(message_id, channel_id)
	if err != nil {
		return fmt.Errorf("Donor_addChannelPost GetPostByDonorIdAndChId err: %v", err)
	}
	if post.PostId != 0 {
		srv.l.Info("пост уже есть в БД, валим!")
		return nil
	}
	// добавили пост в БД
	err = srv.db.AddNewPost(channel_id, message_id, message_id, "")
	if err != nil {
		return err
	}

	// если Media_Group
	if m.ChannelPost.MediaGroupId != nil {
		var postType string
		if len(m.ChannelPost.Photo) > 0 {
			postType = "photo"
		} else if m.ChannelPost.Video.FileId != "" {
			postType = "video"
		} else {
			return fmt.Errorf("Media_Group без photo и video")
		}
		filePath, err := srv.downloadPostMedia(m, postType)
		if err != nil {
			return fmt.Errorf("Donor_addChannelPost downloadPostMedia err: %v", err)
		}
		newmedia := Media{
			Media_group_id:            *m.ChannelPost.MediaGroupId,
			Type_media:                postType,
			fileNameInServer:          filePath,
			Donor_message_id:          message_id,
			Reply_to_donor_message_id: 0,
			Caption:                   "",
			Caption_entities:          m.ChannelPost.CaptionEntities,
			//File_id: // нужно для подтверждения в доноре, позже в вампирах заменяем
			//Reply_to_message_id:  // нужно для подтверждения в доноре, позже в вампирах заменяем
		}
		if postType == "photo" {
			newmedia.File_id = m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId
		} else if postType == "video" {
			newmedia.File_id = m.ChannelPost.Video.FileId
		}
		if m.ChannelPost.ReplyToMessage != nil {
			newmedia.Reply_to_message_id = m.ChannelPost.ReplyToMessage.MessageId
			newmedia.Reply_to_donor_message_id = m.ChannelPost.ReplyToMessage.MessageId
		}
		if m.ChannelPost.Caption != nil {
			newmedia.Caption = *m.ChannelPost.Caption
		}

		srv.MediaCh <- newmedia
		return nil
	}

	// если не Media_Group
	allVampBots, err := srv.db.GetAllVampBots()
	if err != nil {
		return fmt.Errorf("Donor_addChannelPost GetAllVampBots err: %v", err)
	}
	for i, vampBot := range allVampBots {
		if vampBot.ChId == 0 {
			continue
		}
		err := srv.sendChPostAsVamp(vampBot, m)
		if err != nil {
			srv.l.Error("Donor_addChannelPost: sendChPostAsVamp", zap.Error(err))
		}
		srv.l.Info("Donor_addChannelPost", zap.Any("bot index in arr", i), zap.Any("bot ch link", vampBot.ChLink))
		time.Sleep(time.Second * 2)
	}
	srv.l.Info("Donor_addChannelPost: end")

	return nil
}

func (srv *TgService) sendChPostAsVamp(vampBot entity.Bot, m models.Update) error {
	donor_ch_mes_id := m.ChannelPost.MessageId

	if m.ChannelPost.VideoNote != nil {
		//////////////// если кружочек видео
		err := srv.sendChPostAsVamp_VideoNote(vampBot, m)
		return err
	} else if len(m.ChannelPost.Photo) > 0 {
		//////////////// если фото
		err := srv.sendChPostAsVamp_Video_or_Photo(vampBot, m, "photo")
		return err
	} else if m.ChannelPost.Video != nil {
		//////////////// если видео
		err := srv.sendChPostAsVamp_Video_or_Photo(vampBot, m, "video")
		return err
	} else {
		//////////////// если просто текст
		futureMesJson := map[string]any{
			"chat_id": strconv.Itoa(vampBot.ChId),
		}
		if m.ChannelPost.ReplyToMessage != nil {
			replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
			currPost, err := srv.db.GetPostsByDonorIdAndChId_Max(replToDonorChPostId, vampBot.ChId) // тут
			if err != nil {
				return fmt.Errorf("sendChPostAsVamp GetPostByDonorIdAndChId err: %v", err)
			}
			futureMesJson["reply_to_message_id"] = currPost.PostId
		}
		if m.ChannelPost.ReplyMarkup != nil {
			var inlineKeyboardMarkup models.InlineKeyboardMarkup
			mycopy.DeepCopy(m.ChannelPost.ReplyMarkup, &inlineKeyboardMarkup)

			newInlineKeyboardMarkup, err := srv.PrepareReplyMarkup(inlineKeyboardMarkup, vampBot)
			if err != nil {
				return fmt.Errorf("sendChPostAsVamp PrepareReplyMarkup err: %v", err)
			}
			futureMesJson["reply_markup"] = newInlineKeyboardMarkup
		}

		var messText string                            // строка в которую скопируем значение текста поста, тк структуры копируются по ебаной ссылке, и если срезаем часть текста то потом везде так будет
		mycopy.DeepCopy(m.ChannelPost.Text, &messText) // какого хуя в Го структуры копируются по ссылке ?

		if len(m.ChannelPost.Entities) > 0 {
			entities := make([]models.MessageEntity, 0)
			mycopy.DeepCopy(m.ChannelPost.Entities, &entities)

			var newEntities []models.MessageEntity
			var err error

			newEntities, messText, err = srv.PrepareEntities(entities, messText, vampBot)
			if err != nil {
				return fmt.Errorf("sendChPostAsVamp PrepareEntities err: %v", err)
			}
			if newEntities != nil {
				futureMesJson["entities"] = newEntities
			}
		}

		futureMesJson["text"] = messText

		json_data, err := json.Marshal(futureMesJson)
		if err != nil {
			return err
		}
		srv.l.Info("sendChPostAsVamp -> если просто текст -> http.Post", zap.Any("futureMesJson", futureMesJson), zap.Any("string(json_data)", string(json_data)))
		sendVampPostResp, err := http.Post(
			fmt.Sprintf(srv.Cfg.TgEndp, vampBot.Token, "sendMessage"),
			"application/json",
			bytes.NewBuffer(json_data),
		)
		if err != nil {
			return err
		}
		defer sendVampPostResp.Body.Close()
		var cAny struct {
			Ok          bool   `json:"ok"`
			ErrorCode   int    `json:"error_code"`
			Description string `json:"description"`
			Result struct {
				MessageId int `json:"message_id"`
				Caption   string `json:"caption"`
			} `json:"result"`
		}
		if err := json.NewDecoder(sendVampPostResp.Body).Decode(&cAny); err != nil {
			return err
		}
		if cAny.ErrorCode != 0 {
			errMess := fmt.Errorf("sendChPostAsVamp Post ErrorResp: %+v", cAny)
			srv.l.Error(errMess.Error())
		}
		if cAny.Result.MessageId != 0 {
			err = srv.db.AddNewPost(vampBot.ChId, cAny.Result.MessageId, donor_ch_mes_id, cAny.Result.Caption)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (srv *TgService) sendChPostAsVamp_VideoNote(vampBot entity.Bot, m models.Update) error {
	donor_ch_mes_id := m.ChannelPost.MessageId
	futureVideoNoteJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	if m.ChannelPost.ReplyToMessage != nil {
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.db.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_VideoNote: %v", err)
		}
		futureVideoNoteJson["reply_to_message_id"] = strconv.Itoa(currPost.PostId)
	}
	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.Cfg.TgEndp, srv.Cfg.Token, fmt.Sprintf("getFile?file_id=%s", m.ChannelPost.VideoNote.FileId)),
	)
	if err != nil {
		return err
	}
	defer getFilePAthResp.Body.Close()
	var cAny struct {
		Ok     bool `json:"ok"`
		Result struct {
			File_id        string `json:"file_id"`
			File_unique_id string `json:"file_unique_id"`
			File_path      string `json:"file_path"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(getFilePAthResp.Body).Decode(&cAny); err != nil {
		return err
	}
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	srv.l.Info("sendChPostAsVamp_VideoNote: fileNameInServer:", zap.Any("fileNameInServer", fileNameInServer))
	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Cfg.Token, cAny.Result.File_path),
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
		fmt.Sprintf(srv.Cfg.TgEndp, vampBot.Token, "sendVideoNote"),
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
			Caption   string `json:"caption"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return err
	}
	if cAny2.Result.MessageId != 0 {
		err = srv.db.AddNewPost(vampBot.ChId, cAny2.Result.MessageId, donor_ch_mes_id, cAny2.Result.Caption)
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *TgService) sendChPostAsVamp_Video_or_Photo(vampBot entity.Bot, m models.Update, postType string) error {
	donor_ch_mes_id := m.ChannelPost.MessageId
	futureVideoJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	if m.ChannelPost.ReplyToMessage != nil {
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.db.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_Video_or_Photo GetPostByDonorIdAndChId err: %v", err)
		}
		futureVideoJson["reply_to_message_id"] = strconv.Itoa(currPost.PostId)
	}
	if m.ChannelPost.ReplyMarkup != nil {
		var inlineKeyboardMarkup models.InlineKeyboardMarkup
		mycopy.DeepCopy(m.ChannelPost.ReplyMarkup, &inlineKeyboardMarkup)

		newInlineKeyboardMarkup, err := srv.PrepareReplyMarkup(inlineKeyboardMarkup, vampBot)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_Video_or_Photo PrepareReplyMarkup err: %v", err)
		}
		json_data, err := json.Marshal(newInlineKeyboardMarkup)
		if err != nil {
			srv.l.Error("sendChPostAsVamp_Video_or_Photo Marshal err", zap.Error(err), zap.Any("newInlineKeyboardMarkup", newInlineKeyboardMarkup))
		}
		futureVideoJson["reply_markup"] = string(json_data)
	}

	if m.ChannelPost.Caption != nil {
		futureVideoJson["caption"] = *m.ChannelPost.Caption
	}

	if len(m.ChannelPost.CaptionEntities) > 0 {
		entities := make([]models.MessageEntity, len(m.ChannelPost.CaptionEntities))
		mycopy.DeepCopy(m.ChannelPost.CaptionEntities, &entities)

		var newEntities []models.MessageEntity
		var err error

		newEntities, _, err = srv.PrepareEntities(entities, "", vampBot)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp PrepareEntities err: %v", err)
		}
		if newEntities != nil {
			j, _ := json.Marshal(entities)
			futureVideoJson["caption_entities"] = string(j)
		}
	}

	fileId := ""
	if postType == "photo" && len(m.ChannelPost.Photo) > 0 {
		fileId = m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId
	} else if m.ChannelPost.Video != nil {
		fileId = m.ChannelPost.Video.FileId
	}

	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.Cfg.TgEndp, srv.Cfg.Token, fmt.Sprintf("getFile?file_id=%s", fileId)),
	)
	if err != nil {
		return err
	}
	defer getFilePAthResp.Body.Close()
	var cAny struct {
		Ok     bool `json:"ok"`
		Result struct {
			File_id        string `json:"file_id"`
			File_unique_id string `json:"file_unique_id"`
			File_path      string `json:"file_path"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(getFilePAthResp.Body).Decode(&cAny); err != nil {
		return err
	}
	if !cAny.Ok {
		return fmt.Errorf("NOT OK GET %s FILE PATH! _", postType)
	}
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	srv.l.Info("sendChPostAsVamp_VideoNote: fileNameInServer:", zap.Any("fileNameInServer", fileNameInServer))
	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Cfg.Token, cAny.Result.File_path),
		)
		if err != nil {
			srv.l.Error("sendChPostAsVamp_Video_or_Photo: send_ch_post_as_vamp.go:318", zap.Error(err))
			return err
		}
	}

	futureVideoJson[postType] = fmt.Sprintf("@%s", fileNameInServer)

	cf, body, err := files.CreateForm(futureVideoJson)
	if err != nil {
		return err
	}
	method := "sendVideo"
	if postType == "photo" {
		method = "sendPhoto"
	}
	rrres, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgEndp, vampBot.Token, method),
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
			MessageId int    `json:"message_id"`
			Caption   string `json:"caption"`
			Chat struct {
				Id int `json:"id"`
			} `json:"chat"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return err
	}
	if cAny2.Result.MessageId != 0 {
		err = srv.db.AddNewPost(vampBot.ChId, cAny2.Result.MessageId, donor_ch_mes_id, cAny2.Result.Caption)
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *TgService) downloadPostMedia(m models.Update, postType string) (string, error) {
	fileId := ""
	if postType == "photo" {
		fileId = m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId
	} else if m.ChannelPost.Video != nil {
		fileId = m.ChannelPost.Video.FileId
	}
	srv.l.Info(fmt.Sprintf("downloadPostMedia: getting file: %s", fmt.Sprintf(srv.Cfg.TgEndp, srv.Cfg.Token, "getFile?file_id="+fileId)))
	
	GetFileResp, err := srv.GetFile(fileId)
	if err != nil {
		return "", fmt.Errorf("downloadPostMedia GetFile fileId-%s err: %v", fileId, err)
	}
	
	fileNameDir := strings.Split(GetFileResp.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", GetFileResp.Result.File_unique_id, fileNameDir[1])
	
	err = files.DownloadFile(
		fileNameInServer,
		fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Cfg.Token, GetFileResp.Result.File_path),
	)
	if err != nil {
		return "", fmt.Errorf("in method downloadPostMedia[3] err: %s", err)
	}
	return fileNameInServer, nil
}

func (srv *TgService) sendAndDeleteMedia(vampBot entity.Bot, fileNameInServer string, postType string) (string, error) {
	futureJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	futureJson[postType] = fmt.Sprintf("@%s", fileNameInServer)
	
	cf, body, err := files.CreateForm(futureJson)
	if err != nil {
		return "", fmt.Errorf("sendAndDeleteMedia CreateForm err: %v", err)
	}
	method := "sendVideo"
	if postType == "photo" {
		method = "sendPhoto"
	}

	rrres, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgEndp, vampBot.Token, method),
		cf,
		body,
	)
	if err != nil {
		return "", fmt.Errorf("sendAndDeleteMedia Post err: %v", err)
	}
	defer rrres.Body.Close()
	var cAny2 struct {
		Ok     bool `json:"ok"`
		ErrorCode   any `json:"error_code"`
		Description any `json:"description"`
		Result struct {
			MessageId int `json:"message_id"`
			Chat struct {
				Id int `json:"id"`
			} `json:"chat"`
			Video models.Video       `json:"video"`
			Photo []models.PhotoSize `json:"photo"`
		} `json:"result"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return "", fmt.Errorf("sendAndDeleteMedia: Decode: %v", err)
	}
	if !cAny2.Ok {
		return "", fmt.Errorf("sendAndDeleteMedia: NOT OK %s: %+v", method, cAny2)
	}
	DelJson, err := json.Marshal(map[string]any{
		"chat_id":    strconv.Itoa(vampBot.ChId),
		"message_id": strconv.Itoa(cAny2.Result.MessageId),
	})
	if err != nil {
		return "", err
	}
	rrres, err = http.Post(
		fmt.Sprintf(srv.Cfg.TgEndp, vampBot.Token, "deleteMessage"),
		"application/json",
		bytes.NewBuffer(DelJson),
	)
	if err != nil {
		return "", err
	}
	defer rrres.Body.Close()
	var cAny3 struct {
		Ok          bool `json:"ok"`
		Result      any  `json:"result"`
		ErrorCode   any  `json:"error_code"`
		Description any  `json:"description"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny3); err != nil && err != io.EOF {
		return "", err
	}
	if !cAny3.Ok {
		return "", fmt.Errorf("sendAndDeleteMedia: NOT OK deleteMessage : %v", cAny3)
	}
	var fileId string
	if postType == "photo" && len(cAny2.Result.Photo) > 0 {
		fileId = cAny2.Result.Photo[len(cAny2.Result.Photo)-1].FileId
	} else if postType == "video" && cAny2.Result.Video.FileId != "" {
		fileId = cAny2.Result.Video.FileId
	} else {
		return "", fmt.Errorf("sendAndDeleteMedia: no photo no video ;-(")
	}
	return fileId, nil
}

func (s *TgService) sendChPostAsVamp_Media_Group() error {

	mediaArr, ok := s.MediaStore.MediaGroups[StoreKey]
	if !ok {
		return fmt.Errorf("sendChPostAsVamp_Media_Group: not found in MediaStore")
	}

	allVampBots, err := s.db.GetAllVampBots()
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_Media_Group GetAllVampBots err: %v", err)
	}
	if len(allVampBots) == 0 {
		return fmt.Errorf("sendChPostAsVamp_Media_Group GetAllVampBots err: len(allVampBots) == 0")
	}

	for _, vampBot := range allVampBots {
		if vampBot.ChId == 0 {
			continue
		}
		for i, media := range mediaArr {
			fileId, err := s.sendAndDeleteMedia(vampBot, media.fileNameInServer, media.Type_media)
			if err != nil {
				s.l.Error("sendChPostAsVamp_Media_Group: sendAndDeleteMedia err", zap.Any("bot ch link", vampBot.ChLink), zap.Error(err))
			}
			mediaArr[i].File_id = fileId

			if media.Reply_to_donor_message_id != 0 {
				replToDonorChPostId := media.Reply_to_donor_message_id
				currPost, err := s.db.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
				if err != nil {
					s.l.Error("sendChPostAsVamp_Media_Group: GetPostByDonorIdAndChId err", zap.Error(err))
				}
				mediaArr[i].Reply_to_message_id = currPost.PostId
			}

			if len(media.Caption_entities) > 0 {
				entities := make([]models.MessageEntity, 0)
				mycopy.DeepCopy(media.Caption_entities, &entities)

				var newEntities []models.MessageEntity
				var err error

				newEntities, _, err = s.PrepareEntities(entities, "", vampBot)
				if err != nil {
					return fmt.Errorf("sendChPostAsVamp PrepareEntities err: %v", err)
				}
				if newEntities != nil {
					mediaArr[i].Caption_entities = newEntities
				}
			}
		}

		arrsik := make([]models.InputMedia, 0)
		for _, med := range mediaArr {
			nwmd := models.InputMedia{
				Type:            med.Type_media,
				Media:           med.File_id,
				Caption:         med.Caption,
				CaptionEntities: med.Caption_entities,
			}
			ok := MediaInSlice(arrsik, nwmd)
			if !ok {
				arrsik = append(arrsik, nwmd)
			}
		}

		ttttt := map[string]any{
			"chat_id": strconv.Itoa(vampBot.ChId),
			"media":   arrsik,
		}
		if mediaArr[0].Reply_to_message_id != 0 {
			ttttt["reply_to_message_id"] = mediaArr[0].Reply_to_message_id
		}

		MediaJson, err := json.Marshal(ttttt)
		if err != nil {
			s.l.Error("sendChPostAsVamp_Media_Group: json.Marshal(ttttt)", zap.Error(err))
		}
		s.l.Info("sendChPostAsVamp_Media_Group: sending media-group", zap.Any("bot ch link", vampBot.ChLink), zap.Any("map[string]any", ttttt))
		rrresfyhfy, err := http.Post(
			fmt.Sprintf(s.Cfg.TgEndp, vampBot.Token, "sendMediaGroup"),
			"application/json",
			bytes.NewBuffer(MediaJson),
		)
		if err != nil {
			s.l.Error("sendChPostAsVamp_Media_Group: sending media-group err", zap.Error(err))
		}
		defer rrresfyhfy.Body.Close()
		var cAny223 struct {
			Ok          bool   `json:"ok"`
			Description string `json:"description"`
			Result      []struct {
				MessageId int `json:"message_id"`
				Caption   string `json:"caption"`
				Chat      struct {
					Id int `json:"id"`
				} `json:"chat,omitempty"`
				Video models.Video       `json:"video"`
				Photo []models.PhotoSize `json:"photo"`
			} `json:"result,omitempty"`
		}
		if err := json.NewDecoder(rrresfyhfy.Body).Decode(&cAny223); err != nil && err != io.EOF {
			s.l.Error("sendChPostAsVamp_Media_Group: Decode err", zap.Error(err))
		}
		s.l.Info("sendChPostAsVamp_Media_Group: sending media-group response", zap.Any("resp struct", cAny223), zap.Any("bot ch link", vampBot.ChLink))
		for _, v := range cAny223.Result {
			if v.MessageId == 0 {
				continue
			}
			for _, med := range mediaArr {
				err = s.db.AddNewPost(vampBot.ChId, v.MessageId, med.Donor_message_id, v.Caption)
				if err != nil {
					s.l.Error("sendChPostAsVamp_Media_Group: AddNewPost err", zap.Error(err))
				}
			}
		}
	}

	delete(s.MediaStore.MediaGroups, StoreKey)
	return nil
}
