package pg

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

//go:embed schemes/bot.sql
var bots_schema string

//go:embed schemes/post.sql
var posts_schema string

//go:embed schemes/user.sql
var users_schema string

//go:embed schemes/group_link.sql
var group_link_schema string

type(
	DBConfig struct {
		User     string
		Password string
		Database string
		Host     string
		Port     string
	}

	Database struct {
		db   *sql.DB
		l    *zap.Logger
	}
)

func New(config DBConfig, l *zap.Logger) (*Database, error) {
	// databaseURI += "sslmode=disable&default_query_exec_mode=cache_describe&pool_max_conns=10&pool_max_conn_lifetime=1m&pool_max_conn_idle_time=1m"
	databaseURI := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		config.User, config.Password, config.Host, config.Port, config.Database,
	)
	db, err := sql.Open("pgx", databaseURI)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	queries := []string{
		posts_schema,
		bots_schema,
		users_schema,
		group_link_schema,
	}
	for _, v := range queries {
		if _, err := db.Exec(v); err != nil {
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
