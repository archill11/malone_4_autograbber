package tg_service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"myapp/internal/entity"
	"myapp/internal/models"
	"myapp/internal/repository"
	"myapp/pkg/files"
	"myapp/pkg/mycopy"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

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
			// ReplToDonorChId := m.ChannelPost.ReplyToMessage.Chat.Id
			replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
			currPost, err := srv.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
			if err != nil {
				return fmt.Errorf("sendChPostAsVamp (1): %v", err)
			}
			futureMesJson["reply_to_message_id"] = currPost.PostId
		}

		var messText string // строка в которую скопируем значение текста поста, тк структуры копируются по ебаной ссылке, и если срезаем часть текста то потом везде так будет
		if len(m.ChannelPost.Entities) > 0 {
			entities := make([]models.MessageEntity, len(m.ChannelPost.Entities))
			mycopy.DeepCopy(m.ChannelPost.Entities, &entities)
			cutEntities := false
			for i, v := range entities {
				if strings.HasPrefix(v.Url, "http://fake-link") || strings.HasPrefix(v.Url, "fake-link") || strings.HasPrefix(v.Url, "https://fake-link") {
					groupLink, err := srv.As.GetGroupLinkById(vampBot.GroupLinkId)
					if err != nil && !errors.Is(err, repository.ErrNotFound) {
						return err
					}
					srv.l.Info("sendChPostAsVamp -> если просто текст -> entities -> GetGroupLinkById", zap.Any("vampBot", vampBot), zap.Any("groupLink", groupLink))
					if groupLink.Link == "" {
						continue
					}
					if strings.HasPrefix(groupLink.Link, "http://cut-link") || strings.HasPrefix(groupLink.Link, "cut-link") || strings.HasPrefix(groupLink.Link, "https://cut-link") {
						mycopy.DeepCopy(m.ChannelPost.Text, &messText)// какого хуя в Го структуры копируются по ссылке  ??
						messText = strings.Replace(messText, "Переходим по ссылке - ССЫЛКА", "", -1)
						messText = strings.Replace(messText, "👉 РЕГИСТРАЦИЯ ТУТ 👈", "", -1)
						messText = strings.Replace(messText, "🔖 Написать мне 🔖", "", -1)
						cutEntities = true
						break
					}
					entities[i].Url = groupLink.Link
					continue
				}
				urlArr := strings.Split(v.Url, "/")
				for ii, vv := range urlArr {
					if vv == "t.me" && urlArr[ii+1] == "c" {
						refToDonorChPostId, err := strconv.Atoi(urlArr[ii+3])
						if err != nil {
							return err
						}
						currPost, err := srv.As.GetPostByDonorIdAndChId(refToDonorChPostId, vampBot.ChId)
						if err != nil {
							return fmt.Errorf("sendChPostAsVamp (2): %v", err)
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
			if !cutEntities {
				futureMesJson["entities"] = entities
			}
		}

		text_message := m.ChannelPost.Text
		if messText != "" {
			futureMesJson["text"] = messText
		}else{
			futureMesJson["text"] = text_message
		}
		json_data, err := json.Marshal(futureMesJson)
		if err != nil {
			return err
		}
		srv.l.Info("sendChPostAsVamp -> если просто текст -> http.Post", zap.Any("futureMesJson", futureMesJson), zap.Any("string(json_data)", string(json_data)))
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
	donor_ch_mes_id := m.ChannelPost.MessageId
	futureVideoNoteJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	if m.ChannelPost.ReplyToMessage != nil {
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_VideoNote: %v", err)
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

func (srv *TgService) sendChPostAsVamp_Video_or_Photo(vampBot entity.Bot, m models.Update, postType string) error {
	donor_ch_mes_id := m.ChannelPost.MessageId
	futureVideoJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	if m.ChannelPost.ReplyToMessage != nil {
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp_Video_or_Photo (1): %v", err)
		}
		futureVideoJson["reply_to_message_id"] = strconv.Itoa(currPost.PostId)
	}
	if m.ChannelPost.Caption != nil {
		futureVideoJson["caption"] = *m.ChannelPost.Caption
	}
	if len(m.ChannelPost.CaptionEntities) > 0 {
		entities := make([]models.MessageEntity, len(m.ChannelPost.CaptionEntities))
		mycopy.DeepCopy(m.ChannelPost.CaptionEntities, &entities)
		for i, v := range entities {
			if strings.HasPrefix(v.Url, "http://fake-link") || strings.HasPrefix(v.Url, "fake-link") || strings.HasPrefix(v.Url, "https://fake-link") {
				groupLink, err := srv.As.GetGroupLinkById(vampBot.GroupLinkId)
				if err != nil {
					return err
				}
				entities[i].Url = groupLink.Link
				continue
			}
			urlArr := strings.Split(v.Url, "/")
			for ii, vv := range urlArr {
				if len(urlArr) < 4 {
					break
				}
				if vv == "t.me" && urlArr[ii+1] == "c" {
					refToDonorChPostId, err := strconv.Atoi(urlArr[ii+3])
					if err != nil {
						return err
					}
					currPost, err := srv.As.GetPostByDonorIdAndChId(refToDonorChPostId, vampBot.ChId)
					if err != nil {
						return fmt.Errorf("sendChPostAsVamp_Video_or_Photo (2): %v", err)
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

	fileId := ""
	if postType == "photo" && len(m.ChannelPost.Photo) > 0 {
		fileId = m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId
	} else if m.ChannelPost.Video != nil {
		fileId = m.ChannelPost.Video.FileId
	}

	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", fileId)),
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
		return fmt.Errorf("NOT OK GET " + postType + " FILE PATH! _")
	}
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	srv.l.Info("sendChPostAsVamp_VideoNote: fileNameInServer:", zap.Any("fileNameInServer", fileNameInServer))
	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path),
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
		fmt.Sprintf(srv.TgEndp, vampBot.Token, method),
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
			Chat      struct {
				Id int `json:"id"`
			} `json:"chat"`
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

func (srv *TgService) downloadPostMedia(m models.Update, postType string) (string, error) {
	fileId := ""
	if postType == "photo" {
		fileId = m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId
	} else if m.ChannelPost.Video != nil {
		fileId = m.ChannelPost.Video.FileId
	}
	srv.l.Info("downloadPostMedia: getting file: ", zap.Any("url", fmt.Sprintf(srv.TgEndp, srv.Token, "getFile?file_id="+fileId)))
	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", fileId)),
	)
	if err != nil {
		return "", fmt.Errorf("downloadPostMedia: http.Get(1): %s", err)
	}
	defer getFilePAthResp.Body.Close()
	var cAny struct {
		Ok     bool `json:"ok"`
		Result struct {
			File_id        string `json:"file_id"`
			File_unique_id string `json:"file_unique_id"`
			File_path      string `json:"file_path"`
		} `json:"result,omitempty"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(getFilePAthResp.Body).Decode(&cAny); err != nil {
		return "", fmt.Errorf("in method downloadPostMedia[2] err: %s", err)
	}
	if !cAny.Ok {
		fmt.Println("NOT OK GET " + postType + " FILE PATH!")
		return "", fmt.Errorf("NOT OK GET " + postType + " FILE PATH! _")
	}
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	// srv.l.Info("fileNameInServer:", fileNameInServer)
	// srv.l.Info("downloading file:", fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path))
	err = files.DownloadFile(
		fileNameInServer,
		fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path),
	)
	if err != nil {
		srv.l.Error("downloadPostMedia: send_ch_post_as_vamp.go:412", zap.Error(err))
		return "", fmt.Errorf("in method downloadPostMedia[3] err: %s", err)
	}
	return fileNameInServer, nil
}

func (srv *TgService) sendAndDeleteMedia(vampBot entity.Bot, fileNameInServer string, postType string) (string, error) {
	futureJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	futureJson[postType] = fmt.Sprintf("@%s", fileNameInServer)
	// futureJson["disable_notification"] = "true"
	cf, body, err := files.CreateForm(futureJson)
	if err != nil {
		return "", err
	}
	method := "sendVideo"
	if postType == "photo" {
		method = "sendPhoto"
	}
	// srv.l.Info("sending method: ", fmt.Sprintf(srv.TgEndp, vampBot.Token, method))
	rrres, err := http.Post(
		fmt.Sprintf(srv.TgEndp, vampBot.Token, method),
		cf,
		body,
	)
	if err != nil {
		return "", err
	}
	defer rrres.Body.Close()
	var cAny2 struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageId int `json:"message_id"`
			Chat      struct { Id int `json:"id"` } `json:"chat"`
			Video     models.Video                  `json:"video"`
			Photo     []models.PhotoSize            `json:"photo"`
		} `json:"result,omitempty"`
		ErrorCode   any  `json:"error_code,omitempty"`
		Description any  `json:"description,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return "", fmt.Errorf("sendAndDeleteMedia: json.NewDecoder(rrres.Body).Decode(&cAny2): %v", err)
	}
	// srv.l.Info(method, "----resp body:", cAny2)
	// fmt.Println(method, "----resp body:", cAny2)
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
		fmt.Sprintf(srv.TgEndp, vampBot.Token, "deleteMessage"),
		"application/json",
		bytes.NewBuffer(DelJson),
	)
	if err != nil {
		return "", err
	}
	defer rrres.Body.Close()
	var cAny3 struct {
		Ok          bool `json:"ok"`
		Result      any  `json:"result,omitempty"`
		ErrorCode   any  `json:"error_code,omitempty"`
		Description any  `json:"description,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny3); err != nil && err != io.EOF {
		return "", err
	}
	// srv.l.Info("deleteMessage resp body:", cAny3)
	// fmt.Println("-+-deleteMessage resp body:", cAny3)
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

	allVampBots, err := s.As.GetAllVampBots()
	if err != nil {
		s.l.Error("sendChPostAsVamp_Media_Group: s.As.GetAllVampBots", zap.Error(err))
	}
	for _, vampBot := range allVampBots {
		if vampBot.ChId == 0 {
			continue
		}
		for i, media := range mediaArr {
			fileId, err := s.sendAndDeleteMedia(vampBot, media.fileNameInServer, media.Type_media)
			if err != nil {
				s.l.Error("sendChPostAsVamp_Media_Group: s.sendAndDeleteMedia", zap.Error(err), zap.Any("bot ch link", vampBot.ChLink))
			}
			mediaArr[i].File_id = fileId

			// fn replaceReplyMessId
			if media.Reply_to_donor_message_id != 0 {
				replToDonorChPostId := media.Reply_to_donor_message_id
				currPost, err := s.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
				if err != nil {
					s.l.Error("sendChPostAsVamp_Media_Group: service queue (1)", zap.Error(err))
				}
				mediaArr[i].Reply_to_message_id = currPost.PostId
			}
			// fn replaceCaptionEntities
			if len(media.Caption_entities) > 0 {
				entities := make([]models.MessageEntity, len(media.Caption_entities))
				mycopy.DeepCopy(media.Caption_entities, &entities)
				for i, v := range entities {
					if strings.HasPrefix(v.Url, "http://fake-link") || strings.HasPrefix(v.Url, "fake-link") || strings.HasPrefix(v.Url, "https://fake-link") {
						groupLink, err := s.As.GetGroupLinkById(vampBot.GroupLinkId)
						if err != nil {
							s.l.Error("sendChPostAsVamp_Media_Group: GetGroupLinkById", zap.Error(err))
						}
						entities[i].Url = groupLink.Link
						continue
					}
					urlArr := strings.Split(v.Url, "/")
					for ii, vv := range urlArr {
						if len(urlArr) < 4 {
							break
						}
						if vv == "t.me" && urlArr[ii+1] == "c" {
							fmt.Printf("\nэто ссылка на канал %s и пост %s\n", urlArr[ii+2], urlArr[ii+3])
							refToDonorChPostId, err := strconv.Atoi(urlArr[ii+3])
							if err != nil {
								s.l.Error("sendChPostAsVamp_Media_Group: strconv.Atoi (1)", zap.Error(err))
							}
							currPost, err := s.As.GetPostByDonorIdAndChId(refToDonorChPostId, vampBot.ChId)
							if err != nil {
								s.l.Error("sendChPostAsVamp_Media_Group: service queue (2)", zap.Error(err))
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
				mediaArr[i].Caption_entities = entities
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
		s.l.Info("sendChPostAsVamp_Media_Group: sending media-group", zap.Any("map[string]any", ttttt), zap.Any("bot ch link", vampBot.ChLink))
		rrresfyhfy, err := http.Post(
			fmt.Sprintf(s.TgEndp, vampBot.Token, "sendMediaGroup"),
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
				MessageId int `json:"message_id,omitempty"`
				Chat      struct { Id int `json:"id,omitempty"` } `json:"chat,omitempty"`
				Video     models.Video                            `json:"video,omitempty"`
				Photo     []models.PhotoSize                      `json:"photo,omitempty"`
			} `json:"result,omitempty"`
		}
		if err := json.NewDecoder(rrresfyhfy.Body).Decode(&cAny223); err != nil && err != io.EOF {
			s.l.Error("sendChPostAsVamp_Media_Group: json.NewDecoder(rrresfyhfy.Body)", zap.Error(err))
		}
		s.l.Info("sendChPostAsVamp_Media_Group: sending media-group response", zap.Any("resp struct", cAny223), zap.Any("bot ch link", vampBot.ChLink))
		for _, v := range cAny223.Result {
			if v.MessageId == 0 {
				continue
			}
			for _, med := range mediaArr {
				time.Sleep(time.Millisecond * 500)
				err = s.As.AddNewPost(vampBot.ChId, v.MessageId, med.Donor_message_id)
				if err != nil {
					s.l.Error("sendChPostAsVamp_Media_Group: s.As.AddNewPost", zap.Error(err))
				}
			}
		}
	}

	delete(s.MediaStore.MediaGroups, StoreKey)
	return nil
}