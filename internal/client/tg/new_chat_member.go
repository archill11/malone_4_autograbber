package tg

// func (srv *TgClient) HandleNewChatMember(m models.Update) error {
// 	status := m.MyChatMember.NewChatMember.Status
// 	// chat := m.MyChatMember.Chat
// 	// newMemberId := m.MyChatMember.NewChatMember.User.Id
// 	// chatId := chat.Id
// 	// channelTitle := m.MyChatMember.Chat.Title

// 	if !m.MyChatMember.NewChatMember.User.IsBot {
// 		return nil
// 	}
// 	allBots, err := srv.Ts.As.GetAllBots()
// 	if err != nil {
// 		return err
// 	}
// 	ourBot := false
// 	for _, v := range allBots {
// 		if m.MyChatMember.NewChatMember.User.Id == v.Id {
// 			ourBot = true
// 		}
// 	}
// 	if !ourBot {
// 		return nil
// 	}

// 	if status == "administrator" {
// 		err := srv.Ts.NCM_administrator(m)
// 		return err
// 	}

// 	return nil
// }
