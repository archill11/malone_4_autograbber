package tg_service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"myapp/internal/entity"
	"myapp/internal/models"
	"myapp/internal/repository"
	u "myapp/internal/utils"
	"myapp/pkg/mycopy"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

func (srv *TgService) Donor_HandleEditedChannelPost(m models.Update) error {
	chatId := m.EditedChannelPost.Chat.Id
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	srv.l.Info("tgClient: Donor_EditEditedChannelPost", zap.Any("models.Update", m))

	err := srv.Donor_editEditedChannelPost(m)
	if err != nil {
		if err != nil {
			srv.ShowMessClient(chatId, u.ERR_MSG_2+err.Error())
		}
		return err
	}
	return nil
}

func (srv *TgService) Donor_editEditedChannelPost(m models.Update) error {
	// chatId := m.Message.Chat.ID
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	// srv.l.Info("tg_service::AddEditedChannelPost::")

	message_id := m.EditedChannelPost.MessageId

	// Проверка что пост есть уже в базе нужна для того что бы телега не отрпавляла
	// кучу запросов повторно , тк ответ долгий из за рассылки

	// если Media_Group
	if m.EditedChannelPost.MediaGroupId != nil {
		var postType string
		if len(m.EditedChannelPost.Photo) > 0 {
			postType = "photo"
		} else if m.EditedChannelPost.Video.FileId != "" {
			postType = "video"
		} else {
			return fmt.Errorf("Media_Group без photo и video")
		}
		filePath, err := srv.downloadPostMedia(m, postType)
		if err != nil {
			return err
		}
		newmedia := Media{
			Media_group_id:            *m.EditedChannelPost.MediaGroupId,
			Type_media:                postType,
			fileNameInServer:          filePath,
			Donor_message_id:          message_id,
			Reply_to_donor_message_id: 0,
			Caption:                   "",
			Caption_entities:          m.EditedChannelPost.CaptionEntities,
			//File_id: // нужно для подтверждения в доноре, позже в вампирах заменяем
			//Reply_to_message_id:  // нужно для подтверждения в доноре, позже в вампирах заменяем
		}
		if postType == "photo" {
			newmedia.File_id = m.EditedChannelPost.Photo[len(m.EditedChannelPost.Photo)-1].FileId
		} else if postType == "video" {
			newmedia.File_id = m.EditedChannelPost.Video.FileId
		}
		if m.EditedChannelPost.ReplyToMessage != nil {
			newmedia.Reply_to_message_id = m.EditedChannelPost.ReplyToMessage.MessageId
			newmedia.Reply_to_donor_message_id = m.EditedChannelPost.ReplyToMessage.MessageId
		}
		if m.EditedChannelPost.Caption != nil {
			newmedia.Caption = *m.EditedChannelPost.Caption
		}

		srv.MediaCh <- newmedia
		return nil
	}

	// если не Media_Group
	allVampBots, err := srv.As.GetAllVampBots()
	if err != nil {
		return err
	}
	for i, vampBot := range allVampBots {
		if vampBot.ChId == 0 {
			continue
		}
		err := srv.editChPostAsVamp(vampBot, m)
		if err != nil {
			srv.l.Error("Donor_EditChannelPost: editChPostAsVamp", zap.Error(err))
		}
		srv.l.Info("Donor_EditChannelPost", zap.Any("bot index in arr", i), zap.Any("bot ch link", vampBot.ChLink))
		time.Sleep(time.Second * 2)
	}

	return nil
}

func (srv *TgService) editChPostAsVamp(vampBot entity.Bot, m models.Update) error {
	donor_ch_mes_id := m.EditedChannelPost.MessageId
	if strings.ToLower(m.EditedChannelPost.Text) == "deletepost" || strings.ToLower(m.EditedChannelPost.Text) == "delete post" {
		currPost, err := srv.As.GetPostByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp (1): %v", err)
		}
		messageForDelete := currPost.PostId
		err = srv.DeleteMess(vampBot.ChId, messageForDelete)
		if err != nil {
			return err
		}
		return nil
	}
	if m.EditedChannelPost.VideoNote != nil {
		//////////////// если кружочек видео
		return nil
	} else if len(m.EditedChannelPost.Photo) > 0 {
		//////////////// если фото
		return nil
	} else if m.EditedChannelPost.Video != nil {
		//////////////// если видео
		return nil
	} else {
		//////////////// если просто текст
		futureMesJson := map[string]any{
			"chat_id": strconv.Itoa(vampBot.ChId),
		}
		currPost, err := srv.As.GetPostByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
		if err != nil {
			return fmt.Errorf("sendChPostAsVamp (1): %v", err)
		}
		futureMesJson["message_id"] = currPost.PostId

		var messText string // строка в которую скопируем значение текста поста, тк структуры копируются по ебаной ссылке, и если срезаем часть текста то потом везде так будет
		if len(m.EditedChannelPost.Entities) > 0 {
			entities := make([]models.MessageEntity, len(m.EditedChannelPost.Entities))
			mycopy.DeepCopy(m.EditedChannelPost.Entities, &entities)
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
						mycopy.DeepCopy(m.EditedChannelPost.Text, &messText) // какого хуя в Го структуры копируются по ссылке ??
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

		text_message := m.EditedChannelPost.Text
		if messText != "" {
			futureMesJson["text"] = messText
		} else {
			futureMesJson["text"] = text_message
		}
		json_data, err := json.Marshal(futureMesJson)
		if err != nil {
			return err
		}
		srv.l.Info("sendChPostAsVamp -> если просто текст -> http.Post", zap.Any("futureMesJson", futureMesJson), zap.Any("string(json_data)", string(json_data)))
		_, err = http.Post(
			fmt.Sprintf(srv.TgEndp, vampBot.Token, "editMessageText"),
			"application/json",
			bytes.NewBuffer(json_data),
		)
		if err != nil {
			return err
		}
	}
	return nil
}