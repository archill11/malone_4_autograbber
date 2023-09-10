package tg_service

import (
	"encoding/json"
	"fmt"
	"myapp/internal/entity"
	"myapp/internal/models"
	"myapp/pkg/mycopy"
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
	srv.l.Info("tgClient: Donor_HandleEditedChannelPost", zap.Any("models.Update", m))

	err := srv.Donor_editEditedChannelPost(m)
	if err != nil {
		if err != nil {
			srv.SendMessage(chatId, ERR_MSG_2+err.Error())
		}
		return err
	}
	return nil
}

func (srv *TgService) Donor_editEditedChannelPost(m models.Update) error {
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
	allVampBots, err := srv.db.GetAllVampBots()
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

	if strings.ToLower(m.EditedChannelPost.Text) == "deletepost" || strings.ToLower(m.EditedChannelPost.Text) == "delete post" || strings.ToLower(m.EditedChannelPost.Text) == "delete"{
		currPost, err := srv.db.GetPostByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
		if err != nil {
			return fmt.Errorf("editChPostAsVamp GetPostByDonorIdAndChId err: %v", err)
		}
		messageForDelete := currPost.PostId
		err = srv.DeleteMessage(vampBot.ChId, messageForDelete, vampBot.Token)
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
		currPost, err := srv.db.GetPostByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
		if err != nil {
			return fmt.Errorf("editChPostAsVamp GetPostByDonorIdAndChId err: %v", err)
		}
		futureMesJson["message_id"] = currPost.PostId

		var messText string
		mycopy.DeepCopy(m.EditedChannelPost.Text, &messText)

		if len(m.EditedChannelPost.Entities) > 0 {
			entities := make([]models.MessageEntity, 0)
			mycopy.DeepCopy(m.EditedChannelPost.Entities, &entities)

			var newEntities []models.MessageEntity
			var err error

			newEntities, messText, err = srv.PrepareEntities(entities, messText, vampBot)
			if err != nil {
				return fmt.Errorf("editChPostAsVamp PrepareEntities err: %v", err)
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
		srv.l.Info("editChPostAsVamp -> если просто текст -> http.Post", zap.Any("futureMesJson", futureMesJson), zap.Any("string(json_data)", string(json_data)))
		err = srv.EditMessageText(json_data, vampBot.Token)
		if err != nil {
			return err
		}
	}
	return nil
}
