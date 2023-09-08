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

	"go.uber.org/zap"
)

type application struct {
	config *config.Config
	server *api.APIServer
	logger *zap.Logger
	db     *pg.Database
	tgs    *tg_service.TgService
}

func main() {
	var err error
	app := &application{}

	app.config = config.Get()

	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.OutputPaths = []string{"logs/info.log", "stderr"}
	app.logger, err = zapCfg.Build()
	if err != nil {
		log.Fatal("can't init logger", err)
	}
	defer app.logger.Sync()

	app.db, err = pg.New(app.config.Db, app.logger) // БД
	if err != nil {
		log.Fatal(err)
	}
	defer logFnError(app.db.CloseDb)

	app.tgs, err = tg_service.New(app.config.Tg, app.db, app.logger) // Tg Service
	if err != nil {
		log.Fatal(err)
	}

	app.server, err = api.New(app.config.Server, app.logger) // api server
	if err != nil {
		log.Fatal(err)
	}
	app.logger.Info("===============Listenning Server===============")
	go log.Fatal(app.server.Server.Listen(":" + app.config.Server.Port))

	defer func() {
		if err := app.server.Server.Shutdown(); err != nil {
			app.logger.Error("ser.Server.Shutdown()", zap.Error(err))
		}
	}()
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sigint
	app.logger.Info("===============Server stopped===============")
}

func logFnError(fn func() error) {
	if err := fn(); err != nil {
		log.Println(err)
	}
}
