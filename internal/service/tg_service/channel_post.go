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
		srv.SendMessage(fromId, ERR_MSG)
		srv.SendMessage(fromId, err.Error())
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
	
	var okSend int
	var notOkSend int
	var IsDisable int
	var ChId0 int
	refkiMap := map[int]map[string]int{}
	for _, vampBot := range allVampBots {
		botRefka := vampBot.GroupLinkId
		refkiMap[botRefka] = map[string]int{
			"Успешно": 0,
			"Неуспешно": 0,
		}
	}

	for i, vampBot := range allVampBots {
		botRefka := vampBot.GroupLinkId
		if vampBot.ChId == 0 {
			ChId0++
			continue
		}
		if vampBot.IsDisable == 1 {
			IsDisable++
			continue
		}
		srv.l.Info("Donor_addChannelPost", zap.Any("bot index in arr", i), zap.Any("bot ch link", vampBot.ChLink))
		err := srv.sendChPostAsVamp(vampBot, m)
		if err != nil {
			notOkSend++
			_, ok := refkiMap[botRefka]
			if ok {
				refkiMap[botRefka]["Неуспешно"] = refkiMap[botRefka]["Неуспешно"]+1
			}
			srv.l.Error("Donor_addChannelPost: sendChPostAsVamp err", zap.Error(err))
			if strings.Contains(err.Error(), "Bad Request: invalid file_id") {
				srv.SendMessage(channel_id, err.Error())
				srv.l.Info("Donor_addChannelPost: end ERROR")
				return nil
			}
		} else {
			okSend++
			_, ok := refkiMap[botRefka]
			if ok {
				refkiMap[botRefka]["Успешно"] = refkiMap[botRefka]["Успешно"]+1
			}
		}
		time.Sleep(time.Second)
	}
	srv.l.Info("Donor_addChannelPost: end")

	donorBot, _ := srv.db.GetBotInfoByToken(srv.Cfg.Token)


	var reportMess bytes.Buffer
	reportMess.WriteString(fmt.Sprintf("Донор псевдоним: %s\n", srv.Cfg.BotPrefix))
	reportMess.WriteString(fmt.Sprintf("Бот: %s\n", srv.AddAt(donorBot.Username)))
	reportMess.WriteString(fmt.Sprintf("Пост: https://t.me/c/%s/%d\n", strconv.Itoa(channel_id)[4:], message_id))
	reportMess.WriteString(fmt.Sprintf("Всего каналов: %d\n", len(allVampBots)))
	reportMess.WriteString(fmt.Sprintf("Успешно отправлено: %d\n", okSend))
	if notOkSend != 0 {
		reportMess.WriteString(fmt.Sprintf("Неуспешно: %d\n", notOkSend))
	}
	if ChId0 != 0 {
		reportMess.WriteString(fmt.Sprintf("Без подвяз. канала: %d\n", ChId0))
	}
	if IsDisable != 0 {
		reportMess.WriteString(fmt.Sprintf("Отключены от рассылки: %d\n", IsDisable))
	}
	srv.SendMessage(channel_id, reportMess.String())
	if srv.Cfg.BotPrefix != "_test"  { // стата в общий канал
		srv.SendMessageByToken(srv.Cfg.ChForStat, reportMess.String(), srv.Cfg.BotTokenForStat)
	}

	var reportMess2 bytes.Buffer
	for key, val := range refkiMap {
		grLinkName, _ := srv.db.GetGroupLinkById(key)
		reportMess2.WriteString(fmt.Sprintf("Реф: %s\n", grLinkName.Title))
		for k, v := range val {
			reportMess2.WriteString(fmt.Sprintf("%s: %d\n", k, v))
		}
		reportMess2.WriteString("\n")
	}
	if srv.Cfg.BotPrefix != "_test"  { // стата в общий канал
		srv.SendMessageByToken(srv.Cfg.ChForStat, reportMess2.String(), srv.Cfg.BotTokenForStat)
	}

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
	//////////////// если гифка
	if m.ChannelPost.Animation != nil {
		return srv.sendChPostAsVamp_Video_or_Photo(vampBot, m, "animation")
	}
	//////////////// если голосовое
	if m.ChannelPost.Voice != nil {
		return srv.sendChPostAsVamp_Video_or_Photo(vampBot, m, "voice")
	}

	//////////////// если просто текст
	futureMesJson := map[string]any{
		"chat_id": strconv.Itoa(vampBot.ChId),
		"link_preview_options": `{"is_disabled": false}`,
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
	var newInlineKeyboardMarkupForSupergroup models.InlineKeyboardMarkup
	if m.ChannelPost.ReplyMarkup != nil {
		var inlineKeyboardMarkup models.InlineKeyboardMarkup
		mycopy.DeepCopy(m.ChannelPost.ReplyMarkup, &inlineKeyboardMarkup)

		newInlineKeyboardMarkup, err := srv.PrepareReplyMarkup(inlineKeyboardMarkup, vampBot)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_VideoNote PrepareReplyMarkup err: %v", err)
		}
		newInlineKeyboardMarkupForSupergroup = newInlineKeyboardMarkup
		json_data, err := json.Marshal(newInlineKeyboardMarkup)
		if err != nil {
			srv.l.Error("sendChPostAsVamp_VideoNote Marshal err", zap.Error(err), zap.Any("newInlineKeyboardMarkup", newInlineKeyboardMarkup))
		}
		futureVideoNoteJson["reply_markup"] = string(json_data)
	}

	getFileResp, err := srv.GetFile(m.ChannelPost.VideoNote.FileId)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_VideoNote GetFile err: %v", err)
	}
	fileNameDir := strings.Split(getFileResp.Result.File_path, ".")
	fileType := "mp4"
	if len(fileNameDir) > 1 {
		fileType = fileNameDir[1]
	}
	fileNameInServer := fmt.Sprintf("./files/%s.%s", getFileResp.Result.File_unique_id, fileType)
	srv.l.Info(fmt.Sprintf("sendChPostAsVamp_VideoNote: fileNameInServer: %s", fileNameInServer))

	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		filePath := getFileResp.Result.File_path
		filePath = strings.TrimPrefix(filePath, fmt.Sprintf("/var/lib/telegram-bot-api/%s", srv.Cfg.Token))
		tgFileUrl := fmt.Sprintf("%s/file/bot%s/%s", srv.Cfg.TgLocUrl, srv.Cfg.Token, filePath)
		err = files.DownloadFile(fileNameInServer, tgFileUrl)
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

	getChatResp, err := srv.GetChat(vampBot.ChId, vampBot.Token)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_VideoNote GetChat err: %v", err)
	}
	if getChatResp.Result.Type != "supergroup" {
		return nil
	}
	if newInlineKeyboardMarkupForSupergroup.InlineKeyboard == nil {
		return nil
	}
	// for _, inlineKeyboard := range newInlineKeyboardMarkupForSupergroup.InlineKeyboard {
	// 	for _, v := range inlineKeyboard {
	// 		if v.Url == nil && v.Text == "" {
	// 			continue
	// 		}
	// 		err = srv.SendMessageByToken(vampBot.ChId, srv.ChInfoToLinkHTML(*v.Url, v.Text), vampBot.Token)
	// 		if err != nil {
	// 			return fmt.Errorf("sendChPostAsVamp_VideoNote SendMessageByToken for supergroup err: %v", err)
	// 		}
	// 	}
	// }
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
			return fmt.Errorf("sendChPostAsVamp_Video_or_Photo Marshal err: %v", err)
		}
		futureVideoJson["reply_markup"] = string(json_data)
	}

	var caption string
	if m.ChannelPost.Caption != nil {
		caption = *m.ChannelPost.Caption
		futureVideoJson["caption"] = caption
	}
	
	if len(m.ChannelPost.CaptionEntities) > 0 {
		entities := make([]models.MessageEntity, 0)
		mycopy.DeepCopy(m.ChannelPost.CaptionEntities, &entities)

		newEntities, newCaption, err := srv.PrepareEntities(entities, caption, vampBot)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp PrepareEntities err: %v", err)
		}
		if newEntities != nil {
			j, _ := json.Marshal(newEntities)
			futureVideoJson["caption_entities"] = string(j)
		}
		futureVideoJson["caption"] = newCaption
	}

	if m.ChannelPost.HasMediaSpoiler {
		futureVideoJson["has_spoiler"] = "true"
	}

	fileId := ""
	if postType == "photo" && len(m.ChannelPost.Photo) > 0 {
		fileId = m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId
	} else if m.ChannelPost.Video != nil {
		fileId = m.ChannelPost.Video.FileId
		futureVideoJson["width"] = strconv.Itoa(m.ChannelPost.Video.Width)
		futureVideoJson["height"] = strconv.Itoa(m.ChannelPost.Video.Height)
	} else if m.ChannelPost.Animation != nil {
		fileId = m.ChannelPost.Animation.FileId
	} else if m.ChannelPost.Voice != nil {
		fileId = m.ChannelPost.Voice.FileId
	}

	getFileResp, err := srv.GetFile(fileId)
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_Video_or_Photo GetFile fileId-%s err: %v", fileId, err)
	}
	
	fileNameDir := strings.Split(getFileResp.Result.File_path, ".")
	fileType := "mp4"
	if len(fileNameDir) > 1 {
		fileType = fileNameDir[1]
	}
	fileNameInServer := fmt.Sprintf("./files/%s.%s", getFileResp.Result.File_unique_id, fileType)
	srv.l.Info(fmt.Sprintf("sendChPostAsVamp_Video_or_Photo: fileNameInServer: %s", fileNameInServer))

	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		filePath := getFileResp.Result.File_path
		filePath = strings.TrimPrefix(filePath, fmt.Sprintf("/var/lib/telegram-bot-api/%s", srv.Cfg.Token))
		tgFileUrl := fmt.Sprintf("%s/file/bot%s/%s", srv.Cfg.TgLocUrl, srv.Cfg.Token, filePath)
		err = files.DownloadFile(fileNameInServer, tgFileUrl)
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
	} else if postType == "animation" {
		method = "sendAnimation"
	} else if postType == "voice" {
		method = "sendVoice"
	}
	url := fmt.Sprintf(srv.Cfg.TgLocEndp, vampBot.Token, method)
	rrres, err := http.Post(url, cf, body)
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
	} else {
		srv.l.Info(fmt.Sprintf("sendChPostAsVamp_Video_or_Photo: Post resp err: %+v", cAny2))
		return fmt.Errorf("sendChPostAsVamp_Video_or_Photo: Post resp err: %+v", cAny2)
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
	fileType := "mp4"
	if len(fileNameDir) > 1 {
		fileType = fileNameDir[1]
	}
	fileNameInServer := fmt.Sprintf("./files/%s.%s", GetFileResp.Result.File_unique_id, fileType)
	filePath := GetFileResp.Result.File_path
	filePath = strings.TrimPrefix(filePath, fmt.Sprintf("/var/lib/telegram-bot-api/%s", srv.Cfg.Token))
	tgFileUrl := fmt.Sprintf("%s/file/bot%s/%s", srv.Cfg.TgLocUrl, srv.Cfg.Token, filePath)
	err = files.DownloadFile(fileNameInServer, tgFileUrl)
	if err != nil {
		return "", fmt.Errorf("downloadPostMedia DownloadFile err: %v", err)
	}
	srv.l.Info(fmt.Sprintf("downloadPostMedia done, fileNameInServer: %s", fileNameInServer))
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
		fmt.Sprintf(srv.Cfg.TgLocEndp, vampBot.Token, method),
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

func (s *TgService) sendChPostAsVamp_Media_Group(mediaGroupId string) error {
	s.l.Info("sendChPostAsVamp_Media_Group start sending", zap.Any("len s.MediaStore.MediaGroups", len(s.MediaStore.MediaGroups)), zap.Any("s.MediaStore.MediaGroups", s.MediaStore.MediaGroups))
	mediaArr, ok := s.MediaStore.MediaGroups[mediaGroupId]
	if !ok {
		return fmt.Errorf("sendChPostAsVamp_Media_Group: not found in MediaStore")
	}
	s.l.Info("sendChPostAsVamp_Media_Group", zap.Any("len mediaArr", len(mediaArr)), zap.Any("mediaArr", mediaArr))

	allVampBots, err := s.db.GetAllVampBots()
	if err != nil {
		return fmt.Errorf("sendChPostAsVamp_Media_Group GetAllVampBots err: %v", err)
	}
	if len(allVampBots) == 0 {
		return fmt.Errorf("sendChPostAsVamp_Media_Group GetAllVampBots err: len(allVampBots) == 0")
	}
	var okSend int
	var notOkSend int
	var IsDisable int
	var ChId0 int
	for _, vampBot := range allVampBots {
		if vampBot.ChId == 0 {
			ChId0++
			continue
		}
		if vampBot.IsDisable == 1 {
			IsDisable++
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

				newEntities, newText, err := s.PrepareEntities(entities, media.Caption, vampBot)
				if err != nil {
					return fmt.Errorf("sendChPostAsVamp PrepareEntities err: %v", err)
				}
				if newEntities != nil {
					mediaArr[i].Caption_entities = newEntities
				}
				if media.Caption != "" {
					mediaArr[i].Caption = newText
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

		s.l.Info("sendChPostAsVamp_Media_Group: sending media-group", zap.Any("bot ch link", vampBot.ChLink), zap.Any("media_json", mediaJson), zap.Any("bot", vampBot))
		cAny223, err := s.SendMediaGroup(media_json, vampBot.Token)
		if err != nil {
			notOkSend++
			s.l.Error(fmt.Sprintf("sendChPostAsVamp_Media_Group: SendMediaGroup err: %v", err))
		}
		s.l.Info("sendChPostAsVamp_Media_Group SendMediaGroup response", zap.Any("bot ch link", vampBot.ChLink), zap.Any("response", cAny223))
		
		for i, v := range cAny223.Result {
			if i == 0 {
				okSend++
			}
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

	delete(s.MediaStore.MediaGroups, mediaGroupId)
	s.l.Info("sendChPostAsVamp_Media_Group end sending", zap.Any("len s.MediaStore.MediaGroups", len(s.MediaStore.MediaGroups)), zap.Any("s.MediaStore.MediaGroups", s.MediaStore.MediaGroups))

	var reportMess bytes.Buffer
	reportMess.WriteString(fmt.Sprintf("Всего ботов: %d\n", len(allVampBots)))
	reportMess.WriteString(fmt.Sprintf("Успешно отправлено: %d\n", okSend))
	if notOkSend != 0 {
		reportMess.WriteString(fmt.Sprintf("Неуспешно: %d\n", notOkSend))
	}
	if ChId0 != 0 {
		reportMess.WriteString(fmt.Sprintf("Без подвяз. канала: %d\n", ChId0))
	}
	if IsDisable != 0 {
		reportMess.WriteString(fmt.Sprintf("Отключены от рассылки: %d\n", IsDisable))
	}
	donorBot, err := s.db.GetBotInfoByToken(s.Cfg.Token)
	if err != nil {
		s.l.Error(fmt.Errorf("sendChPostAsVamp_Media_Group GetBotInfoByToken err: %v", err).Error())
	}
	s.SendMessage(donorBot.ChId, reportMess.String())

	return nil
}
