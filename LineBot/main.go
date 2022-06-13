package main

import (
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/spf13/viper"
)

var (
	bot *linebot.Client
)

func main() {
	secret, token := readToken()
	_bot, err := linebot.New(secret, token)
	if err != nil {
		panic(err)
	}
	bot = _bot
	MONGODB_CONNSTRING := os.Getenv("MONGODB_CONNSTRING")
	if MONGODB_CONNSTRING == "" {
		MONGODB_CONNSTRING = "mongodb://localhost:27017"
	}
}

func readToken() (string, string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./secret/")
	viper.AddConfigPath(".")
	viper.SetDefault("secret", "<channel secret>")
	viper.SetDefault("token", "<channel access token>")
	_ = viper.ReadInConfig()
	viper.SafeWriteConfig()
	return viper.GetString("secret"), viper.GetString("token")
}

