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
	srv.l.Info("Donor_HandleEditedChannelPost", zap.Any("m.EditedChannelPost", *m.EditedChannelPost), zap.Any("models.Update", m))

	err := srv.Donor_EditEditedChannelPost(m)
	if err != nil {
		srv.SendMessage(chatId, ERR_MSG_2+err.Error())
		return err
	}
	return nil
}

func (srv *TgService) Donor_EditEditedChannelPost(m models.Update) error {
	// message_id := m.EditedChannelPost.MessageId

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
		time.Sleep(time.Second)
	}
	srv.l.Info("Donor_EditChannelPost: end")

	return nil
}

func (srv *TgService) editChPostAsVamp(vampBot entity.Bot, m models.Update) error {
	donor_ch_mes_id := m.EditedChannelPost.MessageId

	if m.EditedChannelPost.VideoNote != nil {
		//////////////// если кружочек видео
		return nil
	} else if len(m.EditedChannelPost.Photo) > 0 {
		//////////////// если фото
		if m.EditedChannelPost.Caption != nil {
			if strings.ToLower(*m.EditedChannelPost.Caption) == "deletepost" || strings.ToLower(*m.EditedChannelPost.Caption) == "delete post" || strings.ToLower(*m.EditedChannelPost.Caption) == "delete"{
				currPosts, err := srv.db.GetPostsByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
				if err != nil {
					return fmt.Errorf("editChPostAsVamp GetPostsByDonorIdAndChId err: %v", err)
				}
				for _, currPost := range currPosts {
					messageForDelete := currPost.PostId
					srv.DeleteMessage(vampBot.ChId, messageForDelete, vampBot.Token)
				}
				return nil
			}

			futureMesJson := map[string]any{
				"chat_id": strconv.Itoa(vampBot.ChId),
			}
			currPosts, err := srv.db.GetPostsByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
			if err != nil {
				return fmt.Errorf("editChPostAsVamp GetPostByDonorIdAndChId err: %v", err)
			}
			for _, currPost := range currPosts {
				if currPost.Caption == "" {
					continue
				}
				futureMesJson["message_id"] = currPost.PostId
				var messText string
				mycopy.DeepCopy(*m.EditedChannelPost.Caption, &messText)
		
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
						futureMesJson["caption_entities"] = newEntities
					}
				}
				futureMesJson["caption"] = messText
				json_data, err := json.Marshal(futureMesJson)
				if err != nil {
					return err
				}
				srv.l.Info("editChPostAsVamp -> если фото -> http.Post", zap.Any("futureMesJson", futureMesJson), zap.Any("string(json_data)", string(json_data)))
				err = srv.EditMessageCaption(json_data, vampBot.Token)
				if err != nil {
					return err
				}
			}
		}
		return nil
	} else if m.EditedChannelPost.Video != nil {
		srv.l.Info("editChPostAsVamp -> Video")
		//////////////// если видео
		if m.EditedChannelPost.Caption != nil {
			if strings.ToLower(*m.EditedChannelPost.Caption) == "deletepost" || strings.ToLower(*m.EditedChannelPost.Caption) == "delete post" || strings.ToLower(*m.EditedChannelPost.Caption) == "delete"{
				currPosts, err := srv.db.GetPostsByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
				if err != nil {
					return fmt.Errorf("editChPostAsVamp GetPostsByDonorIdAndChId err: %v", err)
				}
				srv.l.Info("editChPostAsVamp -> Video", zap.Any("currPosts", currPosts))
				for _, currPost := range currPosts {
					messageForDelete := currPost.PostId
					srv.DeleteMessage(vampBot.ChId, messageForDelete, vampBot.Token)
				}
				return nil
			}

			futureMesJson := map[string]any{
				"chat_id": strconv.Itoa(vampBot.ChId),
			}
			currPosts, err := srv.db.GetPostsByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
			if err != nil {
				return fmt.Errorf("editChPostAsVamp GetPostByDonorIdAndChId err: %v", err)
			}
			for _, currPost := range currPosts {
				if currPost.Caption == "" {
					continue
				}
				futureMesJson["message_id"] = currPost.PostId
				var messText string
				mycopy.DeepCopy(*m.EditedChannelPost.Caption, &messText)
		
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
						futureMesJson["caption_entities"] = newEntities
					}
				}
				futureMesJson["caption"] = messText
				json_data, err := json.Marshal(futureMesJson)
				if err != nil {
					return fmt.Errorf("editChPostAsVamp Marshal err: %v", err)
				}
				srv.l.Info("editChPostAsVamp -> если видео -> http.Post", zap.Any("futureMesJson", futureMesJson), zap.Any("string(json_data)", string(json_data)))
				srv.EditMessageCaption(json_data, vampBot.Token)
			}
		}
		return nil
	} else {
		//////////////// если просто текст
		if strings.ToLower(m.EditedChannelPost.Text) == "deletepost" || strings.ToLower(m.EditedChannelPost.Text) == "delete post" || strings.ToLower(m.EditedChannelPost.Text) == "delete"{
			currPosts, err := srv.db.GetPostsByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
			if err != nil {
				return fmt.Errorf("editChPostAsVamp GetPostByDonorIdAndChId err: %v", err)
			}
			for _, currPost := range currPosts {
				messageForDelete := currPost.PostId
				srv.DeleteMessage(vampBot.ChId, messageForDelete, vampBot.Token)
			}
			return nil
		}

		futureMesJson := map[string]any{
			"chat_id": strconv.Itoa(vampBot.ChId),
		}
		currPosts, err := srv.db.GetPostsByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
		if err != nil {
			return fmt.Errorf("editChPostAsVamp GetPostsByDonorIdAndChId err: %v", err)
		}
		for _, currPost := range currPosts {
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
				return fmt.Errorf("editChPostAsVamp Marshal err: %v", err)
			}
			srv.l.Info("editChPostAsVamp -> если просто текст -> http.Post", zap.Any("futureMesJson", futureMesJson), zap.Any("string(json_data)", string(json_data)))
			srv.EditMessageText(json_data, vampBot.Token)
		}
	}
	return nil
}

func (srv *TgService) deleteChPostAsVamp(vampBot entity.Bot, m models.Update, donor_ch_mes_id int) error {
	currPosts, err := srv.db.GetPostsByDonorIdAndChId(donor_ch_mes_id, vampBot.ChId)
	if err != nil {
		return fmt.Errorf("deleteChPostAsVamp GetPostsByDonorIdAndChId err: %v", err)
	}
	for _, currPost := range currPosts {
		srv.DeleteMessage(vampBot.ChId, currPost.PostId, vampBot.Token)
	}
	return nil
}