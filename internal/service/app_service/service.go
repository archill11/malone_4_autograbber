package service

import (
	pg "myapp/internal/repository/pg"

	"go.uber.org/zap"
)

type AppService struct {
	db *pg.Database
	l  *zap.Logger
}

func New(db *pg.Database, l *zap.Logger) (*AppService, error) {
	s := &AppService{
		db: db,
		l:  l,
	}
	return s, nil
}
