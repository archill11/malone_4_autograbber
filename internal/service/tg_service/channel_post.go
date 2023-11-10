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
	fromId := m.ChannelPost.Chat.Id
	srv.l.Info("Donor_HandleChannelPost", zap.Any("models.Update", m))

	err := srv.Donor_addChannelPost(m)
	if err != nil {
		if err != nil {
			srv.SendMessage(fromId, ERR_MSG + err.Error())
			srv.SendMessage(fromId, err.Error())
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
	srv.db.AddNewPost(channel_id, message_id, message_id, "")

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
			File_name_in_server:       filePath,
			Donor_message_id:          message_id,
			Reply_to_donor_message_id: 0,
			Caption:                   "",
			Caption_entities:          m.ChannelPost.CaptionEntities,
			//File_id:               // нужно для подтверждения в доноре, позже в вампирах заменяем
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
		srv.l.Info("Donor_addChannelPost", zap.Any("bot index in arr", i), zap.Any("bot ch link", vampBot.ChLink))
		err := srv.sendChPostAsVamp(vampBot, m)
		if err != nil {
			srv.l.Error("Donor_addChannelPost: sendChPostAsVamp", zap.Error(err))
		}
		time.Sleep(time.Second)
	}
	srv.l.Info("Donor_addChannelPost: end")

	return nil
}

func (srv *TgService) sendChPostAsVamp(vampBot entity.Bot, m models.Update) error {
	donor_ch_mes_id := m.ChannelPost.MessageId

	//////////////// если кружочек
	if m.ChannelPost.VideoNote != nil {
		return srv.sendChPostAsVamp_VideoNote(vampBot, m)
	}
	//////////////// если фото
	if len(m.ChannelPost.Photo) > 0 {
		return srv.sendChPostAsVamp_Video_or_Photo(vampBot, m, "photo")
	}
	//////////////// если видео
	if m.ChannelPost.Video != nil {
		return srv.sendChPostAsVamp_Video_or_Photo(vampBot, m, "video")
	}

	//////////////// если просто текст
	futureMesJson := map[string]any{
		"chat_id": strconv.Itoa(vampBot.ChId),
		"disable_web_page_preview": true,
	}
	if m.ChannelPost.ReplyToMessage != nil {
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.db.GetPostsByDonorIdAndChId_Max(replToDonorChPostId, vampBot.ChId) // тут
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp GetPostsByDonorIdAndChId_Max err: %v", err)
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

		newEntities, messTextt, err := srv.PrepareEntities(entities, messText, vampBot)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp PrepareEntities err: %v", err)
		}
		messText = messTextt
		if newEntities != nil {
			futureMesJson["entities"] = newEntities
		}
	}
	futureMesJson["text"] = messText

	json_data, err := json.Marshal(futureMesJson)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp Marshal futureMesJson err: %v", err)
	}
	srv.l.Info("sendChPostAsVamp -> если просто текст -> http.Post", zap.Any("futureMesJson", futureMesJson), zap.Any("string(json_data)", string(json_data)))
	sendVampPostResp, err := http.Post(
		fmt.Sprintf(srv.Cfg.TgEndp, vampBot.Token, "sendMessage"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp Post err: %v", err)
	}
	defer sendVampPostResp.Body.Close()
	var cAny struct {
		models.BotErrResp
		Result struct {
			MessageId int    `json:"message_id"`
			Caption   string `json:"caption"`
		} `json:"result"`
	}
	if err := json.NewDecoder(sendVampPostResp.Body).Decode(&cAny); err != nil {
		return fmt.Errorf("sendChPostAsVamp Decode err: %v", err)
	}
	if cAny.ErrorCode != 0 {
		errMess := fmt.Errorf("sendChPostAsVamp Post ErrorResp: %+v", cAny)
		return errMess
	}
	if cAny.Result.MessageId != 0 {
		err = srv.db.AddNewPost(vampBot.ChId, cAny.Result.MessageId, donor_ch_mes_id, cAny.Result.Caption)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp AddNewPost err: %v", err)
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
			return fmt.Errorf("sendChPostAsVamp_VideoNote GetPostByDonorIdAndChId err: %v", err)
		}
		futureVideoNoteJson["reply_to_message_id"] = strconv.Itoa(currPost.PostId)
	}
	getFileResp, err := srv.GetFile(m.ChannelPost.VideoNote.FileId)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_VideoNote GetFile err: %v", err)
	}
	fileNameDir := strings.Split(getFileResp.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", getFileResp.Result.File_unique_id, fileNameDir[1])
	srv.l.Info(fmt.Sprintf("sendChPostAsVamp_VideoNote: fileNameInServer: %s", fileNameInServer))

	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Cfg.Token, getFileResp.Result.File_path),
		)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_VideoNote DownloadFile err: %v", err)
		}
	}
	futureVideoNoteJson["video_note"] = fmt.Sprintf("@%s", fileNameInServer)
	cf, body, err := files.CreateForm(futureVideoNoteJson)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_VideoNote CreateForm err: %v", err)
	}
	cAny2, err := srv.SendVideoNote(body, cf, vampBot.Token)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_VideoNote SendVideoNote err: %v", err)
	}
	if cAny2.Result.MessageId != 0 {
		err = srv.db.AddNewPost(vampBot.ChId, cAny2.Result.MessageId, donor_ch_mes_id, cAny2.Result.Caption)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_VideoNote AddNewPost err: %v", err)
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
		entities := make([]models.MessageEntity, 0)
		mycopy.DeepCopy(m.ChannelPost.CaptionEntities, &entities)

		newEntities, _, err := srv.PrepareEntities(entities, "", vampBot)
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

	cAny, err := srv.GetFile(fileId)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_Video_or_Photo GetFile fileId-%s err: %v", fileId, err)
	}
	
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	srv.l.Info(fmt.Sprintf("sendChPostAsVamp_Video_or_Photo: fileNameInServer: %s", fileNameInServer))

	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Cfg.Token, cAny.Result.File_path),
		)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_Video_or_Photo DownloadFile err: %v", err)
		}
	}
	futureVideoJson[postType] = fmt.Sprintf("@%s", fileNameInServer)

	cf, body, err := files.CreateForm(futureVideoJson)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_Video_or_Photo CreateForm err: %v", err)
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
		return fmt.Errorf("sendChPostAsVamp_Video_or_Photo Post err: %v", err)
	}

	defer rrres.Body.Close()
	var cAny2 models.SendMediaResp
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return fmt.Errorf("sendChPostAsVamp_Video_or_Photo Decode err: %v", err)
	}
	if cAny2.Result.MessageId != 0 {
		err = srv.db.AddNewPost(vampBot.ChId, cAny2.Result.MessageId, donor_ch_mes_id, cAny2.Result.Caption)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_Video_or_Photo AddNewPost err: %v", err)
		}
	}
	return nil
}

func (srv *TgService) downloadPostMedia(m models.Update, postType string) (string, error) {
	var fileId string
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
		return "", fmt.Errorf("downloadPostMedia DownloadFile err: %v", err)
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
	var cAny2 models.SendMediaResp
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return "", fmt.Errorf("sendAndDeleteMedia Decode err: %v", err)
	}
	if cAny2.ErrorCode != 0 {
		return "", fmt.Errorf("sendAndDeleteMedia method-%s errorResp: %+v", method, cAny2)
	}

	err = srv.DeleteMessage(vampBot.ChId, cAny2.Result.MessageId, vampBot.Token)
	if err != nil {
		srv.l.Error(fmt.Sprintf("sendAndDeleteMedia DeleteMessage err: %v", err))
	}

	var fileId string
	if postType == "photo" {
		if len(cAny2.Result.Photo) > 0 {
			fileId = cAny2.Result.Photo[len(cAny2.Result.Photo)-1].FileId
		}
	} else if postType == "video" {
		if cAny2.Result.Video.FileId != "" {
			fileId = cAny2.Result.Video.FileId
		}
	} else {
		return "", fmt.Errorf("sendAndDeleteMedia: no photo, no video :(")
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
			fileId, err := s.sendAndDeleteMedia(vampBot, media.File_name_in_server, media.Type_media)
			if err != nil {
				s.l.Error(fmt.Sprintf("sendChPostAsVamp_Media_Group sendAndDeleteMedia ChLink-%s err", vampBot.ChLink), zap.Error(err))
			}
			mediaArr[i].File_id = fileId

			if media.Reply_to_donor_message_id != 0 {
				replToDonorChPostId := media.Reply_to_donor_message_id
				currPost, err := s.db.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
				if err != nil {
					s.l.Error(fmt.Sprintf("sendChPostAsVamp_Media_Group: GetPostByDonorIdAndChId err: %v", err))
				}
				mediaArr[i].Reply_to_message_id = currPost.PostId
			}

			if len(media.Caption_entities) > 0 {
				entities := make([]models.MessageEntity, 0)
				mycopy.DeepCopy(media.Caption_entities, &entities)

				newEntities, _, err := s.PrepareEntities(entities, "", vampBot)
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

		mediaJson := map[string]any{
			"chat_id": strconv.Itoa(vampBot.ChId),
			"media":   arrsik,
		}
		if mediaArr[0].Reply_to_message_id != 0 {
			mediaJson["reply_to_message_id"] = mediaArr[0].Reply_to_message_id
		}
		media_json, err := json.Marshal(mediaJson)
		if err != nil {
			s.l.Error(fmt.Sprintf("sendChPostAsVamp_Media_Group Marshal err: %v", err))
			continue
		}

		s.l.Info("sendChPostAsVamp_Media_Group: sending media-group", zap.Any("bot ch link", vampBot.ChLink), zap.Any("media_json", mediaJson))
		cAny223, err := s.SendMediaGroup(media_json, vampBot.Token)
		if err != nil {
			s.l.Error(fmt.Sprintf("sendChPostAsVamp_Media_Group: SendMediaGroup err: %v", err))
		}
		s.l.Info("sendChPostAsVamp_Media_Group SendMediaGroup response", zap.Any("bot ch link", vampBot.ChLink), zap.Any("response", cAny223))

		for _, v := range cAny223.Result {
			if v.MessageId == 0 {
				continue
			}
			for _, med := range mediaArr {
				err = s.db.AddNewPost(vampBot.ChId, v.MessageId, med.Donor_message_id, v.Caption)
				if err != nil {
					s.l.Error(fmt.Sprintf("sendChPostAsVamp_Media_Group AddNewPost err: %v", err))
				}
			}
		}
	}

	delete(s.MediaStore.MediaGroups, StoreKey)
	return nil
}
