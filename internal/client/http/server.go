package http

import (
	"myapp/config"
	"myapp/internal/client/tg"
	service "myapp/internal/service/app_service"
	"myapp/pkg/logger"
	"runtime"

	"github.com/gofiber/fiber/v2"
)

type APIServer struct {
	Server  *fiber.App
	service *service.AppService
	tgc     *tg.TgClient
	l       *logger.Logger
	sem     chan struct{}
}

func New(conf config.Config, service *service.AppService, tgc *tg.TgClient, l *logger.Logger) (*APIServer, error) {
	app := fiber.New()
	ser := &APIServer{
		Server:  app,
		service: service,
		tgc:     tgc,
		l:       l,
		sem:     make(chan struct{}, runtime.NumCPU()),
	}

	app.Post("/api/v1/donor/update", ser.donor_Update)
	app.Post("/api/v1/vampire/update", ser.vampire_Update)

	return ser, nil
}
