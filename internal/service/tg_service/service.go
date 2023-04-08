package tg_service

import (
	"myapp/config"
	"myapp/internal/entity"
	"myapp/internal/models"
	as "myapp/internal/service/app_service"
	"myapp/pkg/logger"
	"sync"
)

type TgService struct {
	HostUrl string
	MyPort  string
	TgEndp  string
	Token   string
	As      *as.AppService
	l       *logger.Logger
	LMG     LockMediaGroups
}

type LockMediaGroups struct {
	MediaGroups map[string][]Media
	Mu          sync.Mutex
	MuExecuted  bool
}

type Media struct {
	Media_group_id      string
	Type_media          string
	File_id             string
	Caption             string
	Caption_entities    []models.MessageEntity
	Donor_message_id    int
	Reply_to_message_id int
}

func New(conf config.Config, as *as.AppService, l *logger.Logger) (*TgService, error) {
	s := &TgService{
		HostUrl: conf.MY_URL,
		MyPort:  conf.PORT,
		TgEndp:  conf.TG_ENDPOINT,
		Token:   conf.TOKEN,
		As:      as,
		l:       l,
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

	return s, nil
}
