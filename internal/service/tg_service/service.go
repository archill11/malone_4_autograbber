package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myapp/config"
	"myapp/internal/entity"
	"myapp/internal/models"
	as "myapp/internal/service/app_service"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const StoreKey = "example"

type TgService struct {
	HostUrl    string
	MyPort     string
	TgEndp     string
	Token      string
	As         *as.AppService
	l          *zap.Logger
	MediaCh    chan Media
	MediaStore MediaStore
}

type MediaStore struct {
	MediaGroups map[string][]Media
}

type Media struct {
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

func New(conf config.Config, as *as.AppService, l *zap.Logger) (*TgService, error) {
	s := &TgService{
		HostUrl:    conf.MY_URL,
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

	tgobotResp, err := s.getBotByToken(s.Token)
	if err != nil {
		return s, err
	}
	res := tgobotResp.Result
	bot := entity.NewBot(res.Id, res.UserName, res.FirstName, s.Token, 1)
	err = s.As.AddNewBot(bot.Id, bot.Username, bot.Firstname, bot.Token, bot.IsDonor)
	if err != nil {
		return s, err
	}

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
