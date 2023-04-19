package http

import (
	"encoding/json"
	"fmt"
	"myapp/internal/models"

	"github.com/gofiber/fiber/v2"
)

func (srv *APIServer) donor_Update(c *fiber.Ctx) error {
	m := models.Update{}
	var many any
	if err := c.BodyParser(&m); err != nil {
		srv.l.Err(err)
		return nil
	}
	if err := c.BodyParser(&many); err != nil {
		srv.l.Err(err)
		return nil
	}
	b, err := json.Marshal(many)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\n\ndonor:::message::any::%s\n\n\n", string(b))
	// fmt.Printf("%#v\n", m)

	if m.ChannelPost.Chat.Id != 0 { // on Channel_Post
		go srv.tgc.Donor_HandleChannelPost(m)
		// if err != nil {
		// 	srv.l.Err(err)
		// }
		return nil
	}

	if m.CallbackQuery.From.Id != 0 { // on Callback_Query
		err := srv.tgc.HandleCallbackQuery(m)
		if err != nil {
			srv.l.Err(err)
		}
		return nil
	}

	if m.Message.ReplyToMessage.MessageId != 0 { // on Reply_To_Message
		err := srv.tgc.HandleReplyToMessage(m)
		if err != nil {
			srv.l.Err(err)
		}
		return nil
	}

	if m.Message.Chat.Id != 0 { // on Message
		err := srv.tgc.HandleMessage(m)
		if err != nil {
			srv.l.Err(err)
		}
		return nil
	}

	if m.MyChatMember.NewChatMember.Status != "" { // on New_Chat_Member
		err := srv.tgc.HandleNewChatMember(m)
		if err != nil {
			srv.l.Err(err)
		}
		return nil
	}

	return nil
}

func (srv *APIServer) vampire_Update(c *fiber.Ctx) error {
	m := models.Update{}
	// var many any
	if err := c.BodyParser(&m); err != nil {
		srv.l.Err(err)
		return nil
	}
	// if err := c.BodyParser(&many); err != nil {
	// 	srv.l.Err(err)
	// 	return nil
	// }
	// b, err := json.Marshal(many)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("vampire::message::any::", string(b))
	// fmt.Println("message::::", m)

	// if m.ChannelPost.Chat.Id != 0 { // on Channel_Post
	// 	err := srv.tgc.Vampire_HandleChannelPost(m)
	// 	if err != nil {
	// 		srv.l.Err(err)
	// 	}
	// 	return nil
	// }

	if m.MyChatMember.NewChatMember.Status != "" { // on New_Chat_Member
		err := srv.tgc.HandleNewChatMember(m)
		if err != nil {
			srv.l.Err(err)
		}
		return nil
	}

	return nil
}
