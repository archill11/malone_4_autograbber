package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myapp/internal/models"
	"myapp/internal/repository/pg"
	"myapp/pkg/files"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

var (
	mskLoc, _ = time.LoadLocation("Europe/Moscow")
)

type (
	UpdateConfig struct {
		Offset  int
		Timeout int
		Buffer  int
	}

	TgConfig struct {
		TgUrl           string
		TgEndp          string
		TgLocEndp       string
		TgLocUrl        string
		Token           string
		BotPrefix       string
		ChForStat       int
		BotTokenForStat string
	}

	TgService struct {
		Cfg        TgConfig
		db         *pg.Database
		l          *zap.Logger
		MediaCh    chan Media
		MediaStore MediaStore
	}
)

type (
	MediaStore struct {
		MediaGroups map[string][]Media
	}

	Media struct {
		Media_group_id            string
		Type_media                string
		File_name_in_server       string
		Donor_message_id          int
		Reply_to_donor_message_id int // реплай на сообщение в канале доноре
		Caption                   string
		Caption_entities          []models.MessageEntity
		File_id                   string
		Reply_to_message_id       int // реплай на сообщение в канале вампире
	}
)

func New(conf TgConfig, db *pg.Database, l *zap.Logger) (*TgService, error) {
	s := &TgService{
		Cfg:     conf,
		db:      db,
		l:       l,
		MediaCh: make(chan Media, 10),
		MediaStore: MediaStore{
			MediaGroups: make(map[string][]Media),
		},
	}

	// добавить граббера в базу при первом создании сервиса
	go s.InsertGrabberBot()
	// удаление ненужных файлов
	go s.DeleteOldFiles()
	// удаление потеряных ботов
	go s.DeleteLostBots()
	// уведомление о метке на канале
	go s.AlertScamBots()
	// получение tg Donor updates
	go s.GetTgBotUpdates()
	// когда MediaGroup
	go s.AcceptChPostByAdmin()

	return s, nil
}

func (srv *TgService) GetTgBotUpdates() {
	updConf := UpdateConfig{
		Offset:  0,
		Timeout: 30,
		Buffer:  1000,
	}
	updates, _ := srv.GetUpdatesChan(&updConf, srv.Cfg.Token)
	for update := range updates {
		srv.bot_Update(update)
	}
}

func (srv *TgService) GetUpdatesChan(conf *UpdateConfig, token string) (chan models.Update, chan struct{}) {
	UpdCh := make(chan models.Update, conf.Buffer)
	shutdownCh := make(chan struct{})

	go func() {
		for {
			select {
			case <-shutdownCh:
				close(UpdCh)
				return
			default:
				logMess := fmt.Sprintf(srv.Cfg.TgEndp, token, "getUpdates")
				fmt.Println(logMess)
				updates, err := srv.GetUpdates(conf.Offset, conf.Timeout, token)
				if err != nil {
					srv.l.Error(fmt.Sprintf("GetUpdatesChan GetUpdates err: %v", err))
					srv.l.Error("Failed to get updates, retrying in 4 seconds...")
					time.Sleep(time.Second * 4)
					continue
				}

				for _, update := range updates {
					if update.UpdateId >= conf.Offset {
						conf.Offset = update.UpdateId + 1
						UpdCh <- update
					}
				}
			}
		}
	}()
	return UpdCh, shutdownCh
}

func (srv *TgService) bot_Update(m models.Update) error {
	srv.l.Info("	")
	if m.ChannelPost != nil { // on Channel_Post
		go func() {
			err := srv.Donor_HandleChannelPost(m)
			if err != nil {
				srv.l.Error("Donor_HandleChannelPost err", zap.Error(err))
			}
		}()
		return nil
	}

	if m.EditedChannelPost != nil { // on Edited_Channel_Post
		go func() {
			err := srv.Donor_HandleEditedChannelPost(m)
			if err != nil {
				srv.l.Error("Donor_HandleEditedChannelPost err", zap.Error(err))
			}
		}()
		return nil
	}

	if m.CallbackQuery != nil { // on Callback_Query
		go func() {
			err := srv.HandleCallbackQuery(m)
			if err != nil {
				srv.l.Error("HandleCallbackQuery err", zap.Error(err))
			}
		}()
		return nil
	}

	if m.Message != nil && m.Message.ReplyToMessage != nil { // on Reply_To_Message
		go func() {
			err := srv.HandleReplyToMessage(m)
			if err != nil {
				srv.l.Error("HandleReplyToMessage err", zap.Error(err))
			}
		}()
		return nil
	}

	if m.Message != nil && m.Message.Chat != nil { // on Message
		go func() {
			err := srv.HandleMessage(m)
			if err != nil {
				srv.l.Error("HandleMessage err", zap.Error(err))
			}
		}()
		return nil
	}

	return nil
}

func MediaInSlice(s []models.InputMedia, m models.InputMedia) bool {
	for _, v := range s {
		if v.Media == m.Media {
			return true
		}
	}
	return false
}

func MediaInSlice2(s []Media, m Media) bool {
	for _, v := range s {
		if v.File_name_in_server == m.File_name_in_server {
			return true
		}
	}
	return false
}

func (srv *TgService) DeleteOldFiles() {
	cron := gocron.NewScheduler(mskLoc)
	cron.Every(1).Day().At("02:30").Do(func() {
		err := files.RemoveContentsFromDir("files")
		if err != nil {
			srv.l.Error(fmt.Sprintf("DeleteOldFiles .RemoveContentsFromDir('files') err: %v", err))
		}
		srv.l.Info("DeleteOldFiles At(02:30): ok")
	})
	cron.StartAsync()
}

func (srv *TgService) DeleteLostBots() {
	for{
		time.Sleep(time.Hour*2)

		donorBot, err := srv.db.GetBotInfoByToken(srv.Cfg.Token)
		if err != nil {
			errMess := fmt.Sprintf("DeleteLostBots: GetBotInfoByToken err: %v", err)
			srv.l.Error(errMess)
			srv.SendMessage(donorBot.ChId, errMess)
		}
		if donorBot.Id == 0 {
			errMess := fmt.Sprintf("DeleteLostBots: GetBotInfoByToken err: donorBot.Id == 0")
			srv.l.Error(errMess)
			srv.SendMessage(donorBot.ChId, errMess)
		}

		allBots, err := srv.db.GetAllBots()
		if err != nil {
			errMess := fmt.Sprintf("DeleteLostBots: GetAllBots err: %v", err)
			srv.l.Error(errMess)
			srv.SendMessage(donorBot.ChId, errMess)
		}
		if len(allBots) == 0 {
			errMess := fmt.Sprintf("DeleteLostBots: GetAllBots err: len(allBots) == 0")
			srv.l.Error(errMess)
			srv.SendMessage(donorBot.ChId, errMess)
		}

		for _, bot := range allBots {
			if bot.IsDonor == 1 {
				continue
			}
			resp, err := srv.GetMe(bot.Token)
			if err != nil {
				errMess := fmt.Sprintf("DeleteLostBots: getBotByToken token-%s err: %v", bot.Token, err)
				srv.l.Error(errMess)
				srv.SendMessage(donorBot.ChId, errMess)
			}
			if resp.ErrorCode == 401 && resp.Description == "Unauthorized" {
				srv.db.DeleteBot(bot.Id)

				var mess bytes.Buffer
				mess.WriteString("удален бот без доступа\n")
				mess.WriteString(fmt.Sprintf("бот: @%s | %s\n", bot.Username, bot.Token))
				mess.WriteString(fmt.Sprintf("канал: %d | %s\n", bot.ChId, bot.ChLink))
				logMess := mess.String()

				srv.SendMessage(donorBot.ChId, logMess)
				time.Sleep(time.Second)
			}
		}
	}
}

func (srv *TgService) InsertGrabberBot() {
	time.Sleep(time.Second*4)
	bots, err := srv.db.GetAllBots()
	if err != nil {
		err = fmt.Errorf("InsertGrabberBot GetAllBots err: %v", err)
		srv.l.Error(err.Error())
		return
	}
	if len(bots) > 0 {
		return
	}

	grabberBot, err := srv.GetMe(srv.Cfg.Token)
	if err != nil {
		err = fmt.Errorf("InsertGrabberBot GetMe err: %v", err)
		srv.l.Error(err.Error())
		return
	}
	err = srv.db.AddNewBot(grabberBot.Result.Id, grabberBot.Result.UserName, grabberBot.Result.FirstName, srv.Cfg.Token, 1)
	if err != nil {
		err = fmt.Errorf("InsertGrabberBot AddNewBot err: %v", err)
		srv.l.Error(err.Error())
		return
	}
}

func (srv *TgService) AlertScamBots() {
	for{
		time.Sleep(time.Hour*6)

		donorBot, err := srv.db.GetBotInfoByToken(srv.Cfg.Token)
		if err != nil {
			errMess := fmt.Sprintf("AlertScamBots: GetBotInfoByToken err: %v", err)
			srv.l.Error(errMess)
			srv.SendMessage(donorBot.ChId, errMess)
		}
		if donorBot.Id == 0 {
			errMess := fmt.Sprintf("AlertScamBots: GetBotInfoByToken err: donorBot.Id == 0")
			srv.l.Error(errMess)
			srv.SendMessage(donorBot.ChId, errMess)
		}

		allBots, err := srv.db.GetAllBots()
		if err != nil {
			errMess := fmt.Sprintf("AlertScamBots: GetAllBots err: %v", err)
			srv.l.Error(errMess)
			srv.SendMessage(donorBot.ChId, errMess)
		}
		if len(allBots) == 0 {
			errMess := fmt.Sprintf("AlertScamBots: GetAllBots err: len(allBots) == 0")
			srv.l.Error(errMess)
			srv.SendMessage(donorBot.ChId, errMess)
		}

		for _, bot := range allBots {
			if bot.IsDonor == 1 || bot.ChIsSkam == 1 {
				continue
			}
			resp, err := srv.GetChat(bot.ChId, bot.Token)
			if err != nil {
				errMess := fmt.Sprintf("AlertScamBots: GetChat token-%s err: %v", bot.Token, err)
				srv.l.Error(errMess)
				srv.SendMessage(donorBot.ChId, errMess)
				var logBotMess bytes.Buffer
				logBotMess.WriteString("удален бот\n")
				logBotMess.WriteString(fmt.Sprintf("Донор псевдоним: %s\n", srv.Cfg.BotPrefix))
				logBotMess.WriteString(fmt.Sprintf("%s\n", srv.AddAt(bot.Username)))
				logBotMess.WriteString(fmt.Sprintf("%s\n", bot.Token))
				logBotMess.WriteString(fmt.Sprintf("%s\n", bot.ChLink))
				logBotMess.WriteString(fmt.Sprintf("%d\n", bot.ChId))
				grLink, _ := srv.db.GetGroupLinkById(bot.GroupLinkId)
				logBotMess.WriteString(fmt.Sprintf("group_link: %d, %s - %s\n", bot.GroupLinkId, grLink.Title, grLink.Link))
				srv.SendMessage(donorBot.ChId, logBotMess.String())
				if srv.Cfg.BotPrefix != "_test"  { // стата в общий канал
					srv.SendMessageByToken(srv.Cfg.ChForStat, logBotMess.String(), srv.Cfg.BotTokenForStat)
				}
				// srv.db.DeleteBot(bot.Id)
			}
			if strings.Contains(resp.Result.Description, "this account as a scam or a fake") {
				var mess bytes.Buffer
				mess.WriteString("обнаружен скам на канале\n")
				mess.WriteString(fmt.Sprintf("Донор псевдоним: %s\n", srv.Cfg.BotPrefix))
				mess.WriteString(fmt.Sprintf("бот: @%s | %s\n", bot.Username, bot.Token))
				mess.WriteString(fmt.Sprintf("канал: %s | %d\n", bot.ChLink, bot.ChId))
				logMess := mess.String()

				srv.SendMessage(donorBot.ChId, logMess)
				srv.db.EditBotChIsSkam(bot.Id, 1)
				if srv.Cfg.BotPrefix != "_test"  { // стата в общий канал
					srv.SendMessageByToken(srv.Cfg.ChForStat, mess.String(), srv.Cfg.BotTokenForStat)
				}

				time.Sleep(time.Second)
			}
		}
	}
}

func (srv *TgService) AcceptChPostByAdmin() {
	mediaArr := make([]Media, 0)
	for {
		select {
		case x, ok := <-srv.MediaCh:
			if ok {
				okk := MediaInSlice2(mediaArr, x)
				if !okk {
					mediaArr = append(mediaArr, x)
				}
			} else {
				srv.l.Error("AcceptChPostByAdmin closed!")
				return
			}
		case <-time.After(time.Second * 15):
			if len(mediaArr) == 0 {
				continue
			}
			if len(mediaArr) == 1 {
				srv.l.Error("AcceptChPostByAdmin len(mediaArr) == 1")
				continue
			}

			mediaGroupId := mediaArr[0].Media_group_id
			srv.MediaStore.MediaGroups[mediaGroupId] = mediaArr

			arrsik := make([]models.InputMedia, 0)
			for _, med := range mediaArr {
				nwmd := models.InputMedia{
					Type:            med.Type_media,
					Media:           med.File_id,
					Caption:         med.Caption,
					CaptionEntities: med.Caption_entities,
				}
				ok := MediaInSlice(arrsik, nwmd)
				if !ok {
					arrsik = append(arrsik, nwmd)
				}
			}

			donorBot, err := srv.db.GetBotInfoByToken(srv.Cfg.Token)
			if err != nil {
				srv.l.Error(fmt.Sprintf("AcceptChPostByAdmin: GetBotInfoByToken token-%s err: %v", srv.Cfg.Token, err))
			}

			acceptMess := map[string]any{
				"chat_id": strconv.Itoa(donorBot.ChId),
				"media":   arrsik,
			}
			if mediaArr[0].Reply_to_message_id != 0 {
				acceptMess["reply_to_message_id"] = mediaArr[0].Reply_to_message_id
			}
			media_json, err := json.Marshal(acceptMess)
			if err != nil {
				srv.l.Error(fmt.Sprintf("AcceptChPostByAdmin: json.Marshal(acceptMess) err: %v", err))
			}
			err = srv.sendData(media_json, "sendMediaGroup")
			if err != nil {
				srv.l.Error(fmt.Sprintf("AcceptChPostByAdmin: sendData(sendMediaGroup) err: %v", err))
			}

			cfgVal, _ := srv.db.GetCfgValById("auto-acc-media-gr")
			if cfgVal.Val == "1" {
				m := models.Update{
					CallbackQuery: &models.CallbackQuery{
						From: models.User{Id: 0, UserName: "auto"},
					},
				}
				srv.CQ_accept_ch_post_by_admin(m, mediaGroupId)
			} else {
				media_json, err = json.Marshal(map[string]any{
					"chat_id":      strconv.Itoa(donorBot.ChId),
					"text":         "подтвердите сообщение сверху",
					"reply_markup": fmt.Sprintf(`{ "inline_keyboard" : [[{ "text": "разослать по каналам", "callback_data": "accept_ch_post_%s_by_admin" }]] }`, mediaGroupId),
				})
				if err != nil {
					srv.l.Error(fmt.Sprintf("AcceptChPostByAdmin: Marshal media_json err: %v", err))
				}
				err = srv.sendData(media_json, "sendMessage")
				if err != nil {
					srv.l.Error(fmt.Sprintf("AcceptChPostByAdmin: sendData(sendMessage) err: %v", err))
				}
			}

			mediaArr = mediaArr[0:0]
		}
	}
}