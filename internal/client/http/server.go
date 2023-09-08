package http

import (
	"runtime"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type(
	SerConfig struct {
		Port string
	}

	APIServer struct {
		Server  *fiber.App
		l       *zap.Logger
		sem     chan struct{}
	}
) 

func New(conf SerConfig, l *zap.Logger) (*APIServer, error) {
	app := fiber.New()
	ser := &APIServer{
		Server:  app,
		l:       l,
		sem:     make(chan struct{}, runtime.NumCPU()),
	}

	return ser, nil
}
