package service

import (
	pg "myapp/internal/repository/pg"
	"myapp/pkg/logger"
)

type AppService struct {
	db *pg.Database
	l  *logger.Logger
}

func New(db *pg.Database, l *logger.Logger) (*AppService, error) {
	s := &AppService{
		db: db,
		l:  l,
	}
	return s, nil
}
