package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Status string
const (
	Pending Status = "pending"
	Accepted Status = "accepted"
)

type Contact struct {
	Id uint `json:"id"`
	SenderId uint `json:"sender_id"`
	ReceiverId uint `json:"receiver_id"`

	Status Status `json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

type ContactWithUserAndChat struct {
	Id uint `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ContactId uint `gorm:"column:contact_id" json:"contact_id"`
	ChatId uint `gorm:"column:chat_id" json:"chat_id"`
	ContactEmail string `gorm:"column:contact_email" json:"contact_email"`
	ContactPublicKey []byte `gorm:"column:contact_public_key" json:"contact_public_key"`
}

func GetContactsByUser(db *gorm.DB, userId uint) ([]ContactWithUserAndChat, error) {
	contactsWithUser := make([]ContactWithUserAndChat, 0)
	err := db.Table("contacts AS c").
		Select(`
			c.id, c.created_at,
			u.id AS contact_id, u.email AS contact_email, u.public_key AS contact_public_key,
			(
				SELECT c.id
				FROM chat_users cu
				JOIN chats c ON cu.chat_id = c.id
				WHERE cu.user_id = contact_id
				AND (
					SELECT COUNT(1)
					FROM chat_users cu1
					WHERE cu1.chat_id = cu.chat_id
					AND cu1.user_id = ?
				) > 0
			) AS chat_id
		`, userId).
		Joins(`
			JOIN users u
			ON (c.sender_id = ? AND u.id = c.receiver_id)
			OR (c.receiver_id = ? AND u.id = c.sender_id)
		`, userId, userId).
		Where("c.status = 'accepted'").
		Where("c.sender_id = ? OR c.receiver_id = ?", userId, userId).
		Scan(&contactsWithUser).Error
	if err != nil {
		return nil, err
	}

	return contactsWithUser, nil
}

func (c Contact) ValidateCreation(db *gorm.DB) (error) {
	result := db.Table("contacts").
		Where("contacts.sender_id = ? AND contacts.receiver_id = ?", c.SenderId, c.ReceiverId).
		Or("contacts.sender_id = ? AND contacts.receiver_id = ?", c.ReceiverId, c.SenderId).
		Find(&Contact{})
	if result.RowsAffected > 0 {
		return errors.New("Invite already sent")
	}

	return nil
}
