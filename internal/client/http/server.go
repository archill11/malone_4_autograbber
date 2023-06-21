package http

import (
	"myapp/config"
	"myapp/internal/client/tg"
	service "myapp/internal/service/app_service"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type APIServer struct {
	Server  *fiber.App
	service *service.AppService
	tgc     *tg.TgClient
	l       *zap.Logger
	sem     chan struct{}
}

func New(conf config.Config, service *service.AppService, tgc *tg.TgClient, l *zap.Logger) (*APIServer, error) {
	app := fiber.New()
	ser := &APIServer{
		Server:  app,
		service: service,
		tgc:     tgc,
		l:       l,
		sem:     make(chan struct{}, runtime.NumCPU()),
	}

	// app.Post("/api/v1/donor/update", ser.donor_Update)
	// app.Post("/api/v1/vampire/update", ser.vampire_Update)

	return ser, nil
}
