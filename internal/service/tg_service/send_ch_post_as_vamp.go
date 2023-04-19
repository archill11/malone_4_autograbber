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
	// "sync"
)

func (srv *TgService) sendChPostAsVamp(vampBot entity.Bot, m models.Update) error {
	donor_ch_mes_id := m.ChannelPost.MessageId

	if m.ChannelPost.VideoNote.FileId != "" {
		//////////////// если кружочек видео
		err := srv.sendChPostAsVamp_VideoNote(vampBot, m)
		return err
	} else if len(m.ChannelPost.Photo) > 0 {
		//////////////// если фото
		err := srv.sendChPostAsVamp_Video_or_Photo(vampBot, m, "photo")
		return err
	} else if m.ChannelPost.Video.FileId != "" {
		//////////////// если видео
		err := srv.sendChPostAsVamp_Video_or_Photo(vampBot, m, "video")
		return err
	} else {
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
				fmt.Println("entities:", entities)
				fmt.Println("v:", v)
				if strings.HasPrefix(v.Url, "http://fake-link") || strings.HasPrefix(v.Url, "fake-link") || strings.HasPrefix(v.Url, "https://fake-link") {
					groupLink, err := srv.As.GetGroupLinkById(vampBot.GroupLinkId)
					if err != nil && !errors.Is(err, repository.ErrNotFound) {
						fmt.Println(555)
						return err
					}
					fmt.Println(666)
					if groupLink.Link == "" {
						continue
					}
					entities[i].Url = groupLink.Link
					continue
				}
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
	fmt.Println("fileNameInServer:", fileNameInServer)
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


func (srv *TgService) sendChPostAsVamp_Video_or_Photo(vampBot entity.Bot, m models.Update, postType string) error {

	// var wg sync.WaitGroup
	if m.ChannelPost.MediaGroupId != "" {
		// wg.Add(1)
		go srv.sendChPostAsVamp_Video_or_Photo_MediaGroup(vampBot, m, postType)
		// wg.Wait()
		return nil
	}
	fmt.Println(postType, " !!!!")
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

	fileId := m.ChannelPost.Video.FileId
	if postType == "photo" && len(m.ChannelPost.Photo) > 0 {
		fileId = m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId
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
		fmt.Println("NOT OK GET " + postType +" FILE PATH!")
		return fmt.Errorf("NOT OK GET " + postType +" FILE PATH! _")
	}
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	fmt.Println("fileNameInServer:", fileNameInServer)
	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path),
		)
		if err != nil {
			srv.l.Err("send_ch_post_as_vamp.go:318 ->",err)
			return err
		}
	}
	if postType == "video" {
		futureVideoJson["video"] = fmt.Sprintf("@%s", fileNameInServer)
	}else {
		futureVideoJson["photo"] = fmt.Sprintf("@%s", fileNameInServer)
	}
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
	fmt.Sprintln()
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

func (srv *TgService) sendChPostAsVamp_Video_or_Photo_MediaGroup(vampBot entity.Bot, m models.Update, postType string) error {
	// defer wg.Done()
	fmt.Println(postType, "_MediaGroup !!!!")
	srv.l.Info(postType, "_MediaGroup !!!!")

	donor_ch_mes_id := m.ChannelPost.MessageId
	futurePhotoJson := map[string]string{
		"chat_id": strconv.Itoa(vampBot.ChId),
	}
	fileId := m.ChannelPost.Video.FileId
	if postType == "photo" && len(m.ChannelPost.Photo) > 0 {
		fileId = m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId
	}
	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", fileId)),
	)
	srv.l.Info("getting file: ", fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", fileId)))
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
	srv.l.Info("getFilePAthResp:: ", cAny)
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	fmt.Println("fileNameInServer:", fileNameInServer)
	srv.l.Info("fileNameInServer:", fileNameInServer)
	_, err = os.Stat(fileNameInServer)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("ТАКОГО ФАЙЛА НЕ СУЩЕСТВУЕТ!")
		srv.l.Info("ТАКОГО ФАЙЛА НЕ СУЩЕСТВУЕТ!", fileNameInServer)
		err = files.DownloadFile(
			fileNameInServer,
			fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path),
		)
		srv.l.Info("downloading file:", fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path))
		if err != nil {
			srv.l.Err("send_ch_post_as_vamp.go:412 ->",err)
			return err
		}
	}
	
	futurePhotoJson[postType] = fmt.Sprintf("@%s", fileNameInServer)
	// futurePhotoJson["disable_notification"] = "true"
	cf, body, err := files.CreateForm(futurePhotoJson)
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
	srv.l.Info("sending method: ", fmt.Sprintf(srv.TgEndp, vampBot.Token, method))
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
			Video models.Video `json:"video"`
			Photo []models.PhotoSize `json:"photo"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return err
	}
	srv.l.Info("rrres body:", cAny2)
	PhotoDelJson, err := json.Marshal(map[string]any{
		"chat_id":    strconv.Itoa(vampBot.ChId),
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
		if srv.LMG.MuExecuted {
			srv.LMG.Mu.Unlock()
			srv.LMG.MuExecuted = false
		}
	}()
	fmt.Printf("rrres body:\n%#v\n\n", cAny2)

	newmedia := Media{
		Media_group_id:   m.ChannelPost.MediaGroupId,
		Type_media:       postType,
		Donor_message_id: donor_ch_mes_id,
	}
	if postType == "photo" && len(cAny2.Result.Photo) > 0 {
		newmedia.File_id = cAny2.Result.Photo[len(cAny2.Result.Photo)-1].FileId
	}else if postType == "video" && cAny2.Result.Video.FileId != "" {
		newmedia.File_id = cAny2.Result.Video.FileId
	}else{
		srv.l.Err("no photo no video :(")
		return nil
	}
	if m.ChannelPost.Caption != "" {
		fmt.Println(postType, "_MediaGroup_Caption !!!!")
		srv.l.Info(postType, "_MediaGroup_Caption !!!!")
		newmedia.Caption = m.ChannelPost.Caption
	}
	if m.ChannelPost.ReplyToMessage.MessageId != 0 {
		fmt.Println(postType, "_MediaGroup_ReplyToMessage !!!!")
		srv.l.Info(postType, "_MediaGroup_ReplyToMessage !!!!")
		replToDonorChPostId := m.ChannelPost.ReplyToMessage.MessageId
		currPost, err := srv.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
		if err != nil {
			return err
		}
		newmedia.Reply_to_message_id = currPost.PostId
	}
	if len(m.ChannelPost.CaptionEntities) > 0 {
		fmt.Println(postType, "_MediaGroup_CaptionEntities !!!!")
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
	hashForMapGroupIdAndChId := fmt.Sprintf("%s:%s", m.ChannelPost.MediaGroupId, strconv.Itoa(vampBot.ChId))
	fmt.Println("hashForMapGroupIdAndChId::", hashForMapGroupIdAndChId)
	srv.LMG.MediaGroups[hashForMapGroupIdAndChId] = append(srv.LMG.MediaGroups[hashForMapGroupIdAndChId], newmedia)
	srv.LMG.Mu.Unlock()
	srv.LMG.MuExecuted = false
	time.Sleep(time.Second * 5)
	srv.LMG.Mu.Lock()
	srv.LMG.MuExecuted = true
	medias, ok := srv.LMG.MediaGroups[hashForMapGroupIdAndChId]
	if !ok {
		return nil
	}
	if len(medias) < 2 {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!!!!!!!!!!       len(medias) < 2      !!!!!!!!!!!!!!!!")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		return nil
		// srv.LMG.Mu.Unlock()
		// srv.LMG.MuExecuted = false
		// time.Sleep(time.Second * 5)
		// srv.LMG.Mu.Lock()
		// srv.LMG.MuExecuted = true
	}
	fmt.Println("len(medias):::", len(medias))
	// if !srv.LMG.MuExecuted {
	// 	srv.LMG.Mu.Lock()
	// 	srv.LMG.MuExecuted = true
	// }
	// srv.LMG.MuExecuted = true
	
	medias2, ok := srv.LMG.MediaGroups[hashForMapGroupIdAndChId]
	if !ok {
		fmt.Println("в MediaGroups нет токой группы: ", hashForMapGroupIdAndChId)
		return nil
	}
	fmt.Println("medias2: ", medias2)
	s2 := make([]Media, len(medias2))
	copy(s2, medias2)
	delete(srv.LMG.MediaGroups, hashForMapGroupIdAndChId)
	srv.LMG.Mu.Unlock()
	srv.LMG.MuExecuted = false

	arrsik := make([]models.InputMedia, 0)
	for _, med := range s2 {
		nwmd := models.InputMedia{
			Type:            med.Type_media,
			Media:           med.File_id,
			Caption:         med.Caption,
			CaptionEntities: med.Caption_entities,
		}
		fmt.Println("medial element: ", nwmd)
		arrsik = append(arrsik, nwmd)
	}

	ttttt := map[string]any{
		"chat_id": strconv.Itoa(vampBot.ChId),
		"media":   arrsik,
	}
	if s2[0].Reply_to_message_id != 0 {
		ttttt["reply_to_message_id"] = s2[0].Reply_to_message_id
	}
	
	MediaJson, err := json.Marshal(ttttt)
	if err != nil {
		return err
	}
	fmt.Println("")
	fmt.Println("MediaJson::::",string(MediaJson))
	fmt.Println("")
	rrresfyhfy, err := http.Post(
		fmt.Sprintf(srv.TgEndp, vampBot.Token, "sendMediaGroup"),
		"application/json",
		bytes.NewBuffer(MediaJson),
	)
	srv.l.Info("sending media-group:" , ttttt)
	if err != nil {
		return err
	}
	defer rrresfyhfy.Body.Close()
	var cAny223 struct {
		Ok     bool `json:"ok"`
		Description     string `json:"description"`
		Result []struct {
			MessageId int `json:"message_id,omitempty"`
			Chat      struct {
				Id int `json:"id,omitempty"`
			} `json:"chat,omitempty"`
			Video models.Video `json:"video,omitempty"`
			Photo []models.PhotoSize `json:"photo,omitempty"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrresfyhfy.Body).Decode(&cAny223); err != nil && err != io.EOF {
		return err
	}
	srv.l.Info("sending media-group response: ", cAny223)
	fmt.Printf("cAny223:::::::::::; %#v\n", cAny223) 
	for _, v := range cAny223.Result {
		if v.MessageId != 0 {
			for _, med := range s2 {
				err = srv.As.AddNewPost(vampBot.ChId, v.MessageId, med.Donor_message_id)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
