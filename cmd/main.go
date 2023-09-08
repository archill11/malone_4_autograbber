package main

import (
	"log"
	"myapp/config"
	api "myapp/internal/client/http"
	pg "myapp/internal/repository/pg"
	tg_service "myapp/internal/service/tg_service"
	"os"
	"os/signal"
	"syscall"

	jsoniter "github.com/json-iterator/go"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

func main() {
	config := config.Get()

	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.OutputPaths = []string{"logs/info.log", "stderr"}
	l, err := zapCfg.Build()
	if err != nil {
		log.Fatal("can't init logger", err)
	}
	defer l.Sync()

	json := jsoniter.ConfigFastest

	db, err := pg.New(config, l, json) // БД
	if err != nil {
		log.Fatal(err)
	}
	defer logFnError(db.CloseDb)

	_, err = tg_service.New(config, db, l) // Tg Service для взаимодейсвия с api телеграм
	if err != nil {
		log.Fatal(err)
	}

	ser, err := api.New(config, l) // api server
	if err != nil {
		log.Fatal(err)
	}
	go log.Fatal(ser.Server.Listen(":" + config.PORT))
	l.Info("===============Listenning Server===============")

	defer func() {
		if err := ser.Server.Shutdown(); err != nil {
			l.Error("ser.Server.Shutdown()", zap.Error(err))
		}
	}()
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sigint
	l.Info("===============Server stopped===============")
}

func logFnError(fn func() error) {
	if err := fn(); err != nil {
		log.Println(err)
	}
}
