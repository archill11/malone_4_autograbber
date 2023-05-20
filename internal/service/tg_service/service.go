package tg_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"myapp/config"
	"myapp/internal/entity"
	"myapp/internal/models"
	as "myapp/internal/service/app_service"
	"myapp/pkg/logger"
	"myapp/pkg/mycopy"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TgService struct {
	HostUrl string
	MyPort  string
	TgEndp  string
	Token   string
	As      *as.AppService
	l       *logger.Logger
	LMG     LockMediaGroups
	MediaCh chan Media
}

type LockMediaGroups struct {
	MediaGroups map[string][]Media
	Mu          sync.Mutex
	MuExecuted  bool
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

func New(conf config.Config, as *as.AppService, l *logger.Logger) (*TgService, error) {
	s := &TgService{
		HostUrl: conf.MY_URL,
		MyPort:  conf.PORT,
		TgEndp:  conf.TG_ENDPOINT,
		Token:   conf.TOKEN,
		As:      as,
		l:       l,
		MediaCh: make(chan Media, 10),
		LMG: LockMediaGroups{
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
		for{
			select {
			case x, ok := <-s.MediaCh:
				if ok {
					ok := MediaInSlice2(mediaArr, x)
					if !ok {
						fmt.Printf("Value %v was read.\n", x)
						mediaArr = append(mediaArr, x)
					}
				} else {
					fmt.Println("Channel closed!")
					return
				}
			case <-time.After(time.Second*15):
				if len(mediaArr) == 0 {
					continue
				}
				if len(mediaArr) == 1 {
					s.l.Err("!!!!!!!!!!  len(mediaArr) == 1  !!!!!!!!!!!!!!!!")
					continue
				}
				// TODO
				// разбить на много разных методов
				allVampBots, err := s.As.GetAllVampBots()
				if err != nil {
					s.l.Err(err)
				}
				for _, vampBot := range allVampBots {
					if vampBot.ChId == 0 {
						continue
					}
					for i, media := range mediaArr {
						fileId, err := s.sendAndDeleteMedia(vampBot, media.fileNameInServer, media.Type_media)
						if err != nil {
							s.l.Err(err)
						}
						fmt.Println("---fileId:", fileId)
						mediaArr[i].File_id = fileId

						// fn replaceReplyMessId
						if media.Reply_to_donor_message_id != 0 {
							fmt.Println(media.Type_media, "_MediaGroup_ReplyToMessage !!!!")
							s.l.Info(media.Type_media, "_MediaGroup_ReplyToMessage !!!!")
							replToDonorChPostId := media.Reply_to_donor_message_id
							currPost, err := s.As.GetPostByDonorIdAndChId(replToDonorChPostId, vampBot.ChId)
							if err != nil {
								s.l.Err(err)
							}
							mediaArr[i].Reply_to_message_id = currPost.PostId
						}
						// fn replaceCaptionEntities
						if len(media.Caption_entities) > 0 {
							fmt.Println(media.Type_media, "_MediaGroup_CaptionEntities !!!!")
							entities := make([]models.MessageEntity, len(media.Caption_entities))
							mycopy.DeepCopy(media.Caption_entities, &entities)
							for i, v := range entities {
								if strings.HasPrefix(v.Url, "http://fake-link") || strings.HasPrefix(v.Url, "fake-link") || strings.HasPrefix(v.Url, "https://fake-link") {
									groupLink, err := s.As.GetGroupLinkById(vampBot.GroupLinkId)
									if err != nil {
										s.l.Err(err)
									}
									entities[i].Url = groupLink.Link
									continue
								}
								urlArr := strings.Split(v.Url, "/")
								for ii, vv := range urlArr {
									if len(urlArr) < 4 {
										break
									}
									if vv == "t.me" && urlArr[ii+1] == "c" {
										fmt.Printf("\nэто ссылка на канал %s и пост %s\n", urlArr[ii+2], urlArr[ii+3])
										refToDonorChPostId, err := strconv.Atoi(urlArr[ii+3])
										if err != nil {
											s.l.Err(err)
										}
										currPost, err := s.As.GetPostByDonorIdAndChId(refToDonorChPostId, vampBot.ChId)
										if err != nil {
											s.l.Err(err)
										}
										if vampBot.ChId < 0 {
											urlArr[ii+2] = strconv.Itoa(-vampBot.ChId)
										} else {
											urlArr[ii+2] = strconv.Itoa(vampBot.ChId)
										}
										if urlArr[ii+2][0] == '1' && urlArr[ii+2][1] == '0' && urlArr[ii+2][2] == '0' {
											urlArr[ii+2] = urlArr[ii+2][3:]
										}
										urlArr[ii+3] = strconv.Itoa(currPost.PostId)
										entities[i].Url = strings.Join(urlArr, "/")
									}
								}
							}
							mediaArr[i].Caption_entities = entities
						}
					}

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
							fmt.Println("medial element: ", nwmd)
							arrsik = append(arrsik, nwmd)
						}
					}

					ttttt := map[string]any{
						"chat_id": strconv.Itoa(vampBot.ChId),
						"media":   arrsik,
					}
					if mediaArr[0].Reply_to_message_id != 0 {
						ttttt["reply_to_message_id"] = mediaArr[0].Reply_to_message_id
					}
					
					MediaJson, err := json.Marshal(ttttt)
					if err != nil {
						s.l.Err(err)
					}
					fmt.Println("")
					fmt.Println("MediaJson::::", string(MediaJson))
					fmt.Println("")
					rrresfyhfy, err := http.Post(
						fmt.Sprintf(s.TgEndp, vampBot.Token, "sendMediaGroup"),
						"application/json",
						bytes.NewBuffer(MediaJson),
					)
					s.l.Info("sending media-group:" , ttttt)
					if err != nil {
						s.l.Err(err)
					}
					defer rrresfyhfy.Body.Close()
					var cAny223 struct {
						Ok          bool `json:"ok"`
						Description string `json:"description"`
						Result      []struct {
							MessageId int `json:"message_id,omitempty"`
							Chat      struct {
								Id int `json:"id,omitempty"`
							} `json:"chat,omitempty"`
							Video models.Video `json:"video,omitempty"`
							Photo []models.PhotoSize `json:"photo,omitempty"`
						} `json:"result,omitempty"`
					}
					if err := json.NewDecoder(rrresfyhfy.Body).Decode(&cAny223); err != nil && err != io.EOF {
						s.l.Err(err)
					}
					s.l.Info("sending media-group response: ", cAny223)
					fmt.Printf("cAny223:::::::::::; %#v\n", cAny223)
					for _, v := range cAny223.Result {
						if v.MessageId != 0 {
							for _, med := range mediaArr {
								time.Sleep(time.Millisecond*500)
								err = s.As.AddNewPost(vampBot.ChId, v.MessageId, med.Donor_message_id)
								if err != nil {
									s.l.Err(err)
								}
							}
						}
					}
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