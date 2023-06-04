package tg

import (
	"myapp/config"
	ts "myapp/internal/service/tg_service"

	"go.uber.org/zap"
)

type TgClient struct {
	Token  string
	TgEndp string
	Ts     *ts.TgService
	l      *zap.Logger
}

func New(conf config.Config, ts *ts.TgService, l *zap.Logger) (*TgClient, error) {
	tg := &TgClient{
		Token:  conf.TOKEN,
		TgEndp: conf.TG_ENDPOINT,
		Ts:     ts,
		l:      l,
	}

	return tg, nil
}
