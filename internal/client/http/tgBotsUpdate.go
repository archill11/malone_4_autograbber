package http

// func (srv *APIServer) donor_Update(c *fiber.Ctx) error {
// 	m := models.Update{}
// 	var many any
// 	if err := c.BodyParser(&m); err != nil {
// 		srv.l.Error("donor_Update: c.BodyParser(&m)", zap.Error(err))
// 		return nil
// 	}
// 	if err := c.BodyParser(&many); err != nil {
// 		srv.l.Error("donor_Update: c.BodyParser(&many)", zap.Error(err))
// 		return nil
// 	}
// 	b, err := json.Marshal(many)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	srv.l.Debug("donor_Update: many", zap.Any("body", string(b)))
// 	// fmt.Printf("%#v\n", m)

// 	if m.ChannelPost != nil { // on Channel_Post
// 		err := srv.tgc.Donor_HandleChannelPost(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: Donor_HandleChannelPost(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	if m.CallbackQuery != nil { // on Callback_Query
// 		err := srv.tgc.HandleCallbackQuery(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: HandleCallbackQuery(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	if m.Message != nil && m.Message.ReplyToMessage != nil { // on Reply_To_Message
// 		err := srv.tgc.HandleReplyToMessage(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: HandleReplyToMessage(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	if m.Message != nil && m.Message.Chat != nil { // on Message
// 		err := srv.tgc.HandleMessage(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: HandleMessage(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	if m.MyChatMember != nil && m.MyChatMember.NewChatMember.Status != "" { // on New_Chat_Member
// 		err := srv.tgc.HandleNewChatMember(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: HandleNewChatMember(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	return nil
// }

// func (srv *APIServer) Donor_Update_v2(m models.Update) error {

// 	if m.ChannelPost != nil { // on Channel_Post
// 		err := srv.tgc.Donor_HandleChannelPost(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: Donor_HandleChannelPost(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	if m.CallbackQuery != nil { // on Callback_Query
// 		err := srv.tgc.HandleCallbackQuery(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: HandleCallbackQuery(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	if m.Message != nil && m.Message.ReplyToMessage != nil { // on Reply_To_Message
// 		err := srv.tgc.HandleReplyToMessage(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: HandleReplyToMessage(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	if m.Message != nil && m.Message.Chat != nil { // on Message
// 		err := srv.tgc.HandleMessage(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: HandleMessage(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	if m.MyChatMember != nil && m.MyChatMember.NewChatMember.Status != "" { // on New_Chat_Member
// 		err := srv.tgc.HandleNewChatMember(m)
// 		if err != nil {
// 			srv.l.Error("donor_Update: HandleNewChatMember(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	return nil
// }

// func (srv *APIServer) vampire_Update(c *fiber.Ctx) error {
// 	m := models.Update{}
// 	// var many any
// 	if err := c.BodyParser(&m); err != nil {
// 		srv.l.Error("vampire_Update: c.BodyParser(&m)", zap.Error(err))
// 		return nil
// 	}

// 	if m.MyChatMember != nil && m.MyChatMember.NewChatMember.Status != "" { // on New_Chat_Member
// 		err := srv.tgc.HandleNewChatMember(m)
// 		if err != nil {
// 			srv.l.Error("vampire_Update: HandleNewChatMember(m)", zap.Error(err))
// 		}
// 		return nil
// 	}

// 	return nil
// }
