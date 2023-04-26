package tg_service

import (
	"errors"
	"fmt"
	"myapp/internal/models"
	"myapp/internal/repository"
	"time"
)

func (srv *TgService) Donor_addChannelPost(m models.Update) error {
	// chatId := m.Message.Chat.ID
	// msgText := m.Message.Text
	// userFirstName := m.Message.From.FirstName
	// userUserName := m.Message.From.UserName
	srv.l.Info("tg_service::AddChannelPost::")

	message_id := m.ChannelPost.MessageId
	channel_id := m.ChannelPost.Chat.Id

	post, err := srv.As.GetPostByDonorIdAndChId(message_id, channel_id)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if post.PostId != 0 {
		srv.l.Info("пост уже есть в БД, валим!")
		return nil
	}

	fmt.Println("|_|")

	// добавили пост в БД
	err = srv.As.AddNewPost(channel_id, message_id, message_id)
	if err != nil {
		return err
	}

	// если Media_Group
	if m.ChannelPost.MediaGroupId != "" {
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
			Media_group_id:            m.ChannelPost.MediaGroupId,
			Type_media:                postType,
			fileNameInServer:          filePath,
			Donor_message_id:          message_id,
			Reply_to_donor_message_id: m.ChannelPost.ReplyToMessage.MessageId,
			Caption:                   m.ChannelPost.Caption,
			Caption_entities:          m.ChannelPost.CaptionEntities,
			// File_id + добавляем позже
			// Reply_to_message_id + добавляем позже
		}
		srv.MediaCh <-newmedia
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
			srv.l.Err("sendChPostAsVamp ERR: ", err)
		}
		srv.l.Warn("bot idx:", i)
		time.Sleep(time.Second*3)
	}

	return nil
}
