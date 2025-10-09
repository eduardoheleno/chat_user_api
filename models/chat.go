package model

import (
	"time"

	"gorm.io/gorm"
)

type Chat struct {
	ID uint `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateChat(db *gorm.DB) (*Chat, error) {
	var chat Chat
	err := db.Create(&chat).Error
	if err != nil {
		return nil, err
	}

	return &chat, nil
}
