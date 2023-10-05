package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"myapp/internal/models"
	"myapp/internal/repository/pg"
	"myapp/pkg/files"
	"net/http"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

const StoreKey = "example"

type (
	UpdateConfig struct {
		Offset  int
		Timeout int
		Buffer  int
	}

	TgConfig struct {
		TgEndp      string
		Token       string
		UserbotHost string
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
		fileNameInServer          string
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

	// удаление ненужных файлов
	go s.DeleteOldFiles()

	go s.DeleteLostBots()
	go s.AlertScamBots()

	// получение tg updates Donor
	go func() {
		updConf := UpdateConfig{
			Offset:  0,
			Timeout: 30,
			Buffer:  1000,
		}
		updates, _ := s.GetUpdatesChan(&updConf, s.Cfg.Token)
		for update := range updates {
			s.Donor_Update_v2(update)
		}
	}()

	// когда MediaGroup
	go func() {
		mediaArr := make([]Media, 0)
		for {
			select {
			case x, ok := <-s.MediaCh:
				if ok {
					okk := MediaInSlice2(mediaArr, x)
					if !okk {
						mediaArr = append(mediaArr, x)
					}
				} else {
					s.l.Error("Channel closed!")
					return
				}
			case <-time.After(time.Second * 15):
				if len(mediaArr) == 0 {
					continue
				}
				if len(mediaArr) == 1 {
					s.l.Error("len(mediaArr) == 1")
					continue
				}

				s.MediaStore.MediaGroups[StoreKey] = mediaArr

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

				DonorBot, err := s.db.GetBotInfoByToken(s.Cfg.Token)
				if err != nil {
					s.l.Error("Channel: s.As.GetBotInfoByToken(s.Token)", zap.Error(err))
				}

				acceptMess := map[string]any{
					"chat_id": strconv.Itoa(DonorBot.ChId),
					"media":   arrsik,
				}
				if mediaArr[0].Reply_to_message_id != 0 {
					acceptMess["reply_to_message_id"] = mediaArr[0].Reply_to_message_id
				}
				MediaJson, err := json.Marshal(acceptMess)
				if err != nil {
					s.l.Error("Channel: json.Marshal(acceptMess)", zap.Error(err))
				}
				err = s.sendData(MediaJson, "sendMediaGroup")
				if err != nil {
					s.l.Error("Channel: http.Post(sendMediaGroup)", zap.Error(err))
				}

				acceptMess = map[string]any{
					"chat_id":      strconv.Itoa(DonorBot.ChId),
					"text":         "подтвердите сообщение сверху",
					"reply_markup": `{ "inline_keyboard" : [[{ "text": "разослать по каналам", "callback_data": "accept_ch_post_by_admin" }]] }`,
				}
				MediaJson, err = json.Marshal(acceptMess)
				if err != nil {
					s.l.Error("Channel: json.Marshal(acceptMess) (2)", zap.Error(err))
				}
				err = s.sendData(MediaJson, "sendMessage")
				if err != nil {
					s.l.Error("Channel: http.Post(sendMessage)", zap.Error(err))
				}

				mediaArr = mediaArr[0:0]
			}
		}
	}()

	return s, nil
}

func (ts *TgService) GetUpdatesChan(conf *UpdateConfig, token string) (chan models.Update, chan struct{}) {
	UpdCh := make(chan models.Update, conf.Buffer)
	shutdownCh := make(chan struct{})

	go func() {
		for {
			select {
			case <-shutdownCh:
				close(UpdCh)
				return
			default:
				updates, err := ts.GetUpdates(conf, token)
				if err != nil {
					log.Println("err: ", err)
					log.Println("Failed to get updates, retrying in 3 seconds...")
					time.Sleep(time.Second * 3)
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

func (ts *TgService) GetUpdates(conf *UpdateConfig, token string) ([]models.Update, error) {
	json_data, err := json.Marshal(map[string]any{
		"offset":  conf.Offset,
		"timeout": conf.Timeout,
	})
	if err != nil {
		return []models.Update{}, err
	}
	fmt.Println(
		fmt.Sprintf(ts.Cfg.TgEndp, token, "getUpdates"),
	)
	resp, err := http.Post(
		fmt.Sprintf(ts.Cfg.TgEndp, token, "getUpdates"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return []models.Update{}, err
	}
	defer resp.Body.Close()

	var j models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return []models.Update{}, err
	}

	return j.Result, err
}

func (srv *TgService) Donor_Update_v2(m models.Update) error {
	if m.ChannelPost != nil { // on Channel_Post
		err := srv.Donor_HandleChannelPost(m)
		if err != nil {
			srv.l.Error("donor_Update: Donor_HandleChannelPost(m)", zap.Error(err))
		}
		return nil
	}

	if m.EditedChannelPost != nil { // on Edited_Channel_Post
		err := srv.Donor_HandleEditedChannelPost(m)
		if err != nil {
			srv.l.Error("donor_Update: Donor_HandleEditedChannelPost(m)", zap.Error(err))
		}
		return nil
	}

	if m.CallbackQuery != nil { // on Callback_Query
		err := srv.HandleCallbackQuery(m)
		if err != nil {
			srv.l.Error("donor_Update: HandleCallbackQuery(m)", zap.Error(err))
		}
		return nil
	}

	if m.Message != nil && m.Message.ReplyToMessage != nil { // on Reply_To_Message
		chatId := m.Message.From.Id
		err := srv.HandleReplyToMessage(m)
		if err != nil {
			srv.l.Error("donor_Update: HandleReplyToMessage(m)", zap.Error(err))
			srv.SendMessage(chatId, fmt.Sprintf("%s: %v", ERR_MSG, err))
		}
		return nil
	}

	if m.Message != nil && m.Message.Chat != nil { // on Message
		err := srv.HandleMessage(m)
		if err != nil {
			srv.l.Error("donor_Update: HandleMessage(m)", zap.Error(err))
		}
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
		if v.fileNameInServer == m.fileNameInServer {
			return true
		}
	}
	return false
}

func (ts *TgService) DeleteOldFiles() {
	mskLoc, _ := time.LoadLocation("Europe/Moscow")
	cron := gocron.NewScheduler(mskLoc)
	cron.Every(1).Day().At("02:30").Do(func() {
		err := files.RemoveContentsFromDir("files")
		if err != nil {
			ts.l.Error(fmt.Sprintf("DeleteOldFiles .RemoveContentsFromDir('files') err: %v", err))
		}
		ts.l.Info("DeleteOldFiles At(02:30): ok")
	})
	cron.StartAsync()
}

func (srv *TgService) DeleteLostBots() {
	for{
		time.Sleep(time.Minute*300)

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
			resp, err := srv.getBotByToken(bot.Token)
			if err != nil {
				errMess := fmt.Sprintf("DeleteLostBots: getBotByToken token-%s err: %v", bot.Token, err)
				srv.l.Error(errMess)
				srv.SendMessage(donorBot.ChId, errMess)
			}
			if !resp.Ok && resp.ErrorCode == 401 && resp.Description == "Unauthorized" {
				srv.db.DeleteBot(bot.Id)

				var mess bytes.Buffer
				mess.WriteString("удален бот без доступа\n")
				mess.WriteString(fmt.Sprintf("бот: @%s | %s\n", bot.Username, bot.Token))
				mess.WriteString(fmt.Sprintf("канал: %d | %s\n", bot.ChId, bot.ChLink))
				logMess := mess.String()

				srv.SendMessage(donorBot.ChId, logMess)
				time.Sleep(time.Second * 3)
			}
		}
	}
}

func (srv *TgService) AlertScamBots() {
	for{
		time.Sleep(time.Minute*300)

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
			if bot.IsDonor == 1 {
				continue
			}
			addChannelResp, err := srv.UB_add_channel_flood(bot.ChLink, donorBot.ChId)
			if err != nil {
				errMess := fmt.Sprintf("AlertScamBots: UB_add_channel_flood err: %v", err)
				srv.l.Error(errMess)
				srv.SendMessage(donorBot.ChId, errMess)
			}
			if addChannelResp.Channel.Id != 0 {
				if addChannelResp.Channel.IsScam {
					var mess bytes.Buffer
					mess.WriteString("обнаружен скам на канале\n")
					mess.WriteString(fmt.Sprintf("бот: @%s | %s\n", bot.Username, bot.Token))
					mess.WriteString(fmt.Sprintf("канал: %d | %s\n", bot.ChId, bot.ChLink))
					logMess := mess.String()

					srv.l.Info(logMess)
					srv.SendMessage(donorBot.ChId, logMess)
				}
			}
		}
	}
}