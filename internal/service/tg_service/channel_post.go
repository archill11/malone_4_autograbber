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

	allVampBots, err := srv.As.GetAllVampBots()
	if err != nil {
		return err
	}

 	for i, vampBot := range allVampBots {
		if vampBot.ChId != 0 {
			srv.sendChPostAsVamp(vampBot, m)
		}
		srv.l.Warn("bot idx:", i)
		// fmt.Println("skip", vampBot)
		time.Sleep(time.Second*7)
	}

	return nil
}
