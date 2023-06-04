package tg_service

import (
	"errors"
	"fmt"
	"myapp/internal/models"
	"myapp/internal/repository"
	"time"

	"go.uber.org/zap"
)

func (srv *TgService) Donor_addChannelPost(m models.Update) error {
	// chatId := m.Message.Chat.ID
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	// srv.l.Info("tg_service::AddChannelPost::")

	message_id := m.ChannelPost.MessageId
	channel_id := m.ChannelPost.Chat.Id

	// Проверка что пост есть уже в базе нужна для того что бы телега не отрпавляла 
	// кучу запросов повторно , тк ответ долгий из за рассылки
	post, err := srv.As.GetPostByDonorIdAndChId(message_id, channel_id)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("Donor_addChannelPost: %v", err)
	}
	if post.PostId != 0 {
		srv.l.Info("пост уже есть в БД, валим!")
		return nil
	}

	// добавили пост в БД
	err = srv.As.AddNewPost(channel_id, message_id, message_id)
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
			return err
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
	allVampBots, err := srv.As.GetAllVampBots()
	if err != nil {
		return err
	}
	for i, vampBot := range allVampBots {
		if vampBot.ChId == 0 {
			continue
		}
		err := srv.sendChPostAsVamp(vampBot, m)
		if err != nil {
			srv.l.Error("Donor_addChannelPost: sendChPostAsVamp", zap.Error(err))
		}
		srv.l.Info("Donor_addChannelPost", zap.Any("bot index in arr", i))
		time.Sleep(time.Second * 3)
	}

	return nil
}
