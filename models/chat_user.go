package model

import (
	"time"

	"gorm.io/gorm"
)

type ChatUser struct {
	ID uint `gorm:"primarykey"`
	ChatId uint `json:"chat_id"`
	UserId uint `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func SaveChatUserBatch(chatId uint, userIds []uint, db *gorm.DB) error {
	chatUsers := make([]ChatUser, len(userIds))
	for i, userId := range userIds {
		chatUsers[i] = ChatUser{
			ChatId: chatId,
			UserId: userId,
		}
	}

	chatUserErr := db.Create(&chatUsers).Error
	if chatUserErr != nil {
		return chatUserErr
	}

	return nil
}
