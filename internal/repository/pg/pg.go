package pg

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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

type (
	DBConfig struct {
		User     string
		Password string
		Database string
		Host     string
		Port     string
	}

	Database struct {
		db *pgxpool.Pool
		l  *zap.Logger
	}
)

func New(config DBConfig, l *zap.Logger) (*Database, error) {
	databaseURI := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		config.User, config.Password, config.Host, config.Port, config.Database,
	)
	databaseURI += "?pool_max_conns=10&pool_max_conn_lifetime=1m&pool_max_conn_idle_time=1m"
	db, err := pgxpool.Connect(context.Background(), databaseURI)
	if err != nil {
		return nil, err
	}

	queries := []string{
		posts_schema,
		bots_schema,
		users_schema,
		group_link_schema,
	}
	for _, v := range queries {
		if _, err := db.Exec(context.Background(), v); err != nil {
			fmt.Println("err", v)
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
	s.db.Close()
	return nil
}

func (s *Database) Exec(sql string, arguments ...any) (pgconn.CommandTag, error) {
	return s.db.Exec(context.Background(), sql, arguments...)
}

func (s *Database) QueryRow(sql string, arguments ...any) pgx.Row {
	return s.db.QueryRow(context.Background(), sql, arguments...)
}

func (s *Database) Query(sql string, arguments ...any) (pgx.Rows, error) {
	return s.db.Query(context.Background(), sql, arguments...)
}
