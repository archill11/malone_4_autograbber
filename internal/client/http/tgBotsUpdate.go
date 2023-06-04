package http

import (
	"encoding/json"
	"fmt"
	"myapp/internal/models"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (srv *APIServer) donor_Update(c *fiber.Ctx) error {
	m := models.Update{}
	var many any
	if err := c.BodyParser(&m); err != nil {
		srv.l.Error("donor_Update: c.BodyParser(&m)", zap.Error(err))
		return nil
	}
	if err := c.BodyParser(&many); err != nil {
		srv.l.Error("donor_Update: c.BodyParser(&many)", zap.Error(err))
		return nil
	}
	b, err := json.Marshal(many)
	if err != nil {
		fmt.Println(err)
	}

	srv.l.Debug("donor_Update: many", zap.Any("body", string(b)))
	// fmt.Printf("%#v\n", m)

	if m.ChannelPost.Chat.Id != 0 { // on Channel_Post
		err := srv.tgc.Donor_HandleChannelPost(m)
		if err != nil {
			srv.l.Error("donor_Update: Donor_HandleChannelPost(m)", zap.Error(err))
		}
		return nil
	}

	if m.CallbackQuery.From.Id != 0 { // on Callback_Query
		err := srv.tgc.HandleCallbackQuery(m)
		if err != nil {
			srv.l.Error("donor_Update: HandleCallbackQuery(m)", zap.Error(err))
		}
		return nil
	}

	if m.Message.ReplyToMessage.MessageId != 0 { // on Reply_To_Message
		err := srv.tgc.HandleReplyToMessage(m)
		if err != nil {
			srv.l.Error("donor_Update: HandleReplyToMessage(m)", zap.Error(err))
		}
		return nil
	}

	if m.Message.Chat.Id != 0 { // on Message
		err := srv.tgc.HandleMessage(m)
		if err != nil {
			srv.l.Error("donor_Update: HandleMessage(m)", zap.Error(err))
		}
		return nil
	}

	if m.MyChatMember.NewChatMember.Status != "" { // on New_Chat_Member
		err := srv.tgc.HandleNewChatMember(m)
		if err != nil {
			srv.l.Error("donor_Update: HandleNewChatMember(m)", zap.Error(err))
		}
		return nil
	}

	return nil
}

func (srv *APIServer) vampire_Update(c *fiber.Ctx) error {
	m := models.Update{}
	// var many any
	if err := c.BodyParser(&m); err != nil {
		srv.l.Error("vampire_Update: c.BodyParser(&m)", zap.Error(err))
		return nil
	}

	if m.MyChatMember.NewChatMember.Status != "" { // on New_Chat_Member
		err := srv.tgc.HandleNewChatMember(m)
		if err != nil {
			srv.l.Error("vampire_Update: HandleNewChatMember(m)", zap.Error(err))
		}
		return nil
	}

	return nil
}
