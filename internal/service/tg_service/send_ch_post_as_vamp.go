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
	fileId := m.ChannelPost.Video.FileId
	if postType == "photo" {
		fileId = m.ChannelPost.Photo[len(m.ChannelPost.Photo)-1].FileId
	}
	srv.l.Info("getting file: ", fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", fileId)))
	fmt.Println("getting file: ", fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", fileId)))
	getFilePAthResp, err := http.Get(
		fmt.Sprintf(srv.TgEndp, srv.Token, fmt.Sprintf("getFile?file_id=%s", fileId)),
	)
	if err != nil {
		return "", fmt.Errorf("in method downloadPostMedia err: %s", err)
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
		fmt.Println("NOT OK GET " + postType +" FILE PATH!")
		return "", fmt.Errorf("NOT OK GET " + postType +" FILE PATH! _")
	}
	srv.l.Info("getFilePAthResp:: ", cAny)
	fileNameDir := strings.Split(cAny.Result.File_path, ".")
	fileNameInServer := fmt.Sprintf("./files/%s.%s", cAny.Result.File_unique_id, fileNameDir[1])
	srv.l.Info("fileNameInServer:", fileNameInServer)
	srv.l.Info("downloading file:", fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path))
	fmt.Println("downloading file:", fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path))
	err = files.DownloadFile(
		fileNameInServer,
		fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", srv.Token, cAny.Result.File_path),
	)
	if err != nil {
		srv.l.Err("send_ch_post_as_vamp.go:412 ->",err)
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
	srv.l.Info("sending method: ", fmt.Sprintf(srv.TgEndp, vampBot.Token, method))
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
			Chat struct {
				Id int `json:"id"`
			} `json:"chat"`
			Video models.Video `json:"video"`
			Photo []models.PhotoSize `json:"photo"`
		} `json:"result,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny2); err != nil && err != io.EOF {
		return "", err
	}
	srv.l.Info(method, "----resp body:", cAny2)
	// fmt.Println(method, "----resp body:", cAny2)
	if !cAny2.Ok {
		return "", fmt.Errorf("NOT OK " + method + " :", cAny2)
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
		Result      any `json:"result,omitempty"`
		ErrorCode   any `json:"error_code,omitempty"`
		Description any `json:"description,omitempty"`
	}
	if err := json.NewDecoder(rrres.Body).Decode(&cAny3); err != nil && err != io.EOF {
		return "", err
	}
	srv.l.Info("deleteMessage resp body:", cAny3)
	// fmt.Println("-+-deleteMessage resp body:", cAny3)
	if !cAny3.Ok {
		return "", fmt.Errorf("NOT OK deleteMessage :", cAny3)
	}
	var fileId string
	if postType == "photo" && len(cAny2.Result.Photo) > 0 {
		fileId = cAny2.Result.Photo[len(cAny2.Result.Photo)-1].FileId
	}else if postType == "video" && cAny2.Result.Video.FileId != "" {
		fileId = cAny2.Result.Video.FileId
	}else{
		return "", fmt.Errorf("no photo no video ;-(")
	}
	return fileId, nil
}

