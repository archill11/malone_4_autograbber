package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"myapp/config"
	"myapp/internal/models"
	as "myapp/internal/service/app_service"
	u "myapp/internal/utils"
	"myapp/pkg/files"
	"net/http"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

const StoreKey = "example"

type TgService struct {
	// HostUrl    string
	MyPort     string
	TgEndp     string
	Token      string
	As         *as.AppService
	l          *zap.Logger
	MediaCh    chan Media
	MediaStore MediaStore
}

type (
	UpdateConfig struct {
		Offset  int
		Timeout int
		Buffer  int
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

func New(conf config.Config, as *as.AppService, l *zap.Logger) (*TgService, error) {
	s := &TgService{
		// HostUrl:    conf.MY_URL,
		MyPort:     conf.PORT,
		TgEndp:     conf.TG_ENDPOINT,
		Token:      conf.TOKEN,
		As:         as,
		l:          l,
		MediaCh:    make(chan Media, 10),
		MediaStore: MediaStore{
			MediaGroups: make(map[string][]Media),
		},
	}

	// tgobotResp, err := s.getBotByToken(s.Token)
	// if err != nil {
	// 	return s, err
	// }
	// res := tgobotResp.Result
	// bot := entity.NewBot(res.Id, res.UserName, res.FirstName, s.Token, 1)
	// err = s.As.AddNewBot(bot.Id, bot.Username, bot.Firstname, bot.Token, bot.IsDonor)
	// if err != nil {
	// 	return s, err
	// }

	// удаление ненужных файлов
	go func() {
		mskLoc, _ := time.LoadLocation("Europe/Moscow")
		cron := gocron.NewScheduler(mskLoc)
		cron.Every(1).Day().At("02:30").Do(func(){
			err := files.RemoveContentsFromDir("files")
			if err != nil {
				s.l.Error(fmt.Sprintf("files.RemoveContentsFromDir('files') err: %v", err))
			}
			s.l.Info("cron.Every(1).Day().At(02:30)")
		})
		cron.StartAsync()
	}()

	// получение tg updates Donor
	go func() {
		updConf := UpdateConfig{
			Offset: 0,
			Timeout: 30,
			Buffer: 1000,
		}
		updates, _ := s.GetUpdatesChan(&updConf, s.Token)
		for update := range updates {
			s.Donor_Update_v2(update)
		}
	}()

	// получение tg updates Vampires
	// go func() {
	// 	var noChannelBotsLen int
	// 	for {
	// 		if noChannelBotsLen > 0 {
	// 			time.Sleep(time.Minute*10)
	// 			continue
	// 		}
	// 		noChannelBots, err := s.As.GetAllNoChannelBots()
	// 		if err != nil {
	// 			s.l.Error("Channel: s.As.GetAllNoChannelBots()", zap.Error(err))
	// 		}
	// 		noChannelBotsLen = len(noChannelBots)
	// 		for _, v := range noChannelBots {
	// 			go func(v entity.Bot){
	// 				updConf := UpdateConfig{
	// 					Offset: 0,
	// 					Timeout: 30,
	// 					Buffer: 1000,
	// 				}
	// 				updates, shutdownCh := s.GetUpdatesChan(&updConf, v.Token)
	// 				for update := range updates {
	// 					closeUpdates, _ := s.Vapmire_Update_v2(update)
	// 					if closeUpdates {
	// 						shutdownCh<- struct{}{}
	// 						s.l.Info("Channel: shutdownCh<- struct{}. Закрыли канал обновлений вампира", zap.Any("bot token", v.Token))
	// 						noChannelBotsLen--
	// 					}
	// 				}
	// 			}(v)
	// 		}
	// 		time.Sleep(time.Minute*10)
	// 	}
	// }()

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
		
				DonorBot, err := s.As.GetBotInfoByToken(s.Token)
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
				_, err = http.Post(
					fmt.Sprintf(s.TgEndp, s.Token, "sendMediaGroup"),
					"application/json",
					bytes.NewBuffer(MediaJson),
				)
				if err != nil {
					s.l.Error("Channel: http.Post(sendMediaGroup)", zap.Error(err))
				}

				acceptMess = map[string]any{
					"chat_id": strconv.Itoa(DonorBot.ChId),
					"text":    "подтвердите сообщение сверху",
					"reply_markup": `{ "inline_keyboard" : [[{ "text": "разослать по каналам", "callback_data": "accept_ch_post_by_admin" }]] }`,
				}
				MediaJson, err = json.Marshal(acceptMess)
				if err != nil {
					s.l.Error("Channel: json.Marshal(acceptMess) (2)", zap.Error(err))
				}
				err = s.sendData(MediaJson)
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
		fmt.Sprintf(ts.TgEndp, token, "getUpdates"),
	)
	resp, err := http.Post(
		fmt.Sprintf(ts.TgEndp, token, "getUpdates"),
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return  []models.Update{}, err
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
			srv.ShowMessClient(chatId, fmt.Sprintf("%s: %v", u.ERR_MSG, err))
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
