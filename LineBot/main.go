package main

import (
	"LineBot/model/dto"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/spf13/viper"
)

var (
	bot *linebot.Client
	db  dto.IConnection
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
	//sql
	_db, err := (&dto.MongoConnection{}).Connect(MONGODB_CONNSTRING)
	if err != nil {
		panic(err)
	}
	db = _db

	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/callback", callback)
	}
	r.RedirectFixedPath = true
	r.Run(":8080")
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

func callback(c *gin.Context) {
	events, err := bot.ParseRequest(c.Request)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			c.Status(400)
		} else {
			c.Status(500)
		}
		return
	}
	if err != nil {
		log.Fatalln(err)
	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				db.Insert(dto.MessageModel{
					UserID:  event.Source.UserID,
					Context: message.Text,
					Type:    "text",
					Time:    event.Timestamp})

				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
					log.Print(err)
				}
			case *linebot.StickerMessage:
				db.Insert(dto.MessageModel{
					UserID:  event.Source.UserID,
					Context: message.StickerID,
					Type:    "sticker",
					Time:    event.Timestamp})
				replyMessage := fmt.Sprintf(
					"sticker id is %s, stickerResourceType is %s", message.StickerID, message.StickerResourceType)
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func wrapResponse(c *gin.Context, context any, err error) {
	var r = struct {
		Context any    `json:"context"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Context: context,
		Status:  "ok", // 預設狀態為ok
		Message: "",
	}

	if err != nil {
		r.Context = nil
		r.Status = "failed"     // 若出現任何err，狀態改為failed
		r.Message = err.Error() // Message回傳錯誤訊息
	}

	c.JSON(http.StatusOK, r)
}
