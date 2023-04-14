package pg

import (
	"database/sql"
	_ "embed"
	"fmt"
	"myapp/config"
	"myapp/pkg/logger"

	_ "github.com/jackc/pgx/v4/stdlib"
)

//go:embed schemes/bots_schema.sql
var bots_schema string

//go:embed schemes/posts_schema.sql
var posts_schema string

//go:embed schemes/users_schema.sql
var users_schema string

//go:embed schemes/group_link_schema.sql
var group_link_schema string

// Database - хранилище заказов.
type Database struct {
	db *sql.DB
	l  *logger.Logger
}

func New(config config.Config, l *logger.Logger) (*Database, error) {
	databaseURI := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		config.PG_USER, config.PG_PASSWORD, config.PG_HOST, "5432", config.PG_DATABASE,
	)
	db, err := sql.Open("pgx", databaseURI)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	if err := db.Ping(); err != nil { // проверка что есть подключеие к БД
		return nil, err
	}
	queries := []string{
		posts_schema,
		bots_schema,
		users_schema,
		group_link_schema,
	}
	for _, v := range queries {
		if _, err := db.Exec(v); err != nil { //создаем таблицы в БД
			return nil, err
		}
	}
	storage := &Database{
		db: db,
		l:  l,
	}
	return storage, nil
}

// CloseDb Метод закрывает соединение с БД
func (s *Database) CloseDb() error {
	return s.db.Close()
}
