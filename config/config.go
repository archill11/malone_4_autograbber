package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// MY_URL      string
	TG_ENDPOINT string
	TOKEN       string
	PORT        string
	PG_USER     string
	PG_PASSWORD string
	PG_DATABASE string
	PG_HOST     string
}

func Get() Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	var c Config

	// c.MY_URL      = os.Getenv("MY_URL")
	c.TG_ENDPOINT = os.Getenv("TG_ENDPOINT")
	c.TOKEN       = os.Getenv("BOT_TOKEN")
	c.PORT        = os.Getenv("APP_PORT")
	c.PG_USER     = os.Getenv("PG_USER")
	c.PG_PASSWORD = os.Getenv("PG_PASSWORD")
	c.PG_DATABASE = os.Getenv("PG_DATABASE")
	c.PG_HOST     = os.Getenv("PG_HOST")

	// c.MY_URL      = ""
	// c.TG_ENDPOINT = "https://api.telegram.org/bot%s/%s"
	// c.TOKEN       = ""
	// c.PORT        = ""
	// c.PG_USER     = ""
	// c.PG_PASSWORD = ""
	// c.PG_DATABASE = ""
	// c.PG_HOST     = ""

	return c
}
