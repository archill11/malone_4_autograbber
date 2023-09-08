package http

import (
	"myapp/config"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type APIServer struct {
	Server  *fiber.App
	l       *zap.Logger
	sem     chan struct{}
}

func New(conf config.Config, l *zap.Logger) (*APIServer, error) {
	app := fiber.New()
	ser := &APIServer{
		Server:  app,
		l:       l,
		sem:     make(chan struct{}, runtime.NumCPU()),
	}

	return ser, nil
}
