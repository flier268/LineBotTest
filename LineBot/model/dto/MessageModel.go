package dto

import "time"

type MessageModel struct {
	UserID  string
	Context string
	Type    string
	Time    time.Time
}
