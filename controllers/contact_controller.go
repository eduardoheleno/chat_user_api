package controller

import (
	"context"
	"encoding/json"
	model "nossochat_api/models"
	"strconv"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ContactCreateBody struct {
	SenderId uint `json:"sender_id"`
	ReceiverId uint `json:"receiver_id"`
	ReceiverEmail string `json:"receiver_email"`
}

type ContactController interface {
	Create(ctx *gin.Context)
	AcceptInvite(ctx *gin.Context)
}

type contactController struct {
	db *gorm.DB
	rabbit *amqp.Channel
	redis *redis.Client
}

type GetPublicKeyWithChatAndMessagesBody struct {
	ContactID uint `json:"contact_id"`
}

func NewContactController(db *gorm.DB, rabbit *amqp.Channel, redis *redis.Client) ContactController {
	return &contactController{db, rabbit, redis}
}

func (c *contactController) Create(ctx *gin.Context) {
	var contactCreateBody ContactCreateBody
	bindErr := ctx.ShouldBindBodyWithJSON(&contactCreateBody)
	if bindErr != nil {
		ctx.JSON(400, gin.H{"message": "Bad request"})
		return
	}

	var contact model.Contact
	contact.Status = model.Pending
	contact.SenderId = contactCreateBody.SenderId
	contact.ReceiverId = contactCreateBody.ReceiverId

	if err := contact.ValidateCreation(c.db); err != nil {
		ctx.JSON(400, gin.H{"message": err.Error()})
		return
	}

	tx := c.db.Begin()
	err := tx.Create(&contact).Error
	if err != nil {
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	result := map[string]interface{}{}
	tx.Table("contacts").
		Select(`
			contacts.*,
			u1.email as sender_email,
			u2.email as receiver_email
		`).
		Joins("right join users u1 on u1.id = contacts.sender_id").
		Joins("right join users u2 on u2.id = contacts.receiver_id").
		Where("contacts.id = ?", contact.Id).
		Scan(&result)
	result["type"] = "Invite"
	result["target_id"] = contact.ReceiverId

	resultJson, parseErr := json.Marshal(result)
	if parseErr != nil {
		tx.Rollback()
		ctx.JSON(500, gin.H{"message": "Couldn't parse invite data"})
		return
	}

	userKey := strconv.FormatUint(uint64(contact.ReceiverId), 10)
	redisCtx := context.Background()
	nodeHash, redisErr := c.redis.Get(redisCtx, userKey).Result()
	if redisErr == nil {
		rabbitErr := c.rabbit.Publish(
			"",
			nodeHash,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType: "application/json",
				Body: resultJson,
			},
		)
		if rabbitErr != nil {
			tx.Rollback()
			ctx.JSON(500, gin.H{"message": "Can't communicate via RabbitMQ"})
			return
		}
	}

	commitErr := tx.Commit().Error
	if commitErr != nil {
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	ctx.JSON(201, result)
}

func (c *contactController) AcceptInvite(ctx *gin.Context) {
	idContact := ctx.Param("idContact")
	authId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	var contact model.Contact
	if queryErr := c.db.Where("id = ?", idContact).
		First(&contact).
		Error; queryErr != nil {
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	if contact.Status == "accepted" {
		ctx.JSON(400, gin.H{"message": "Invite already accepted"})
		return
	}

	tx := c.db.Begin()

	contact.Status = "accepted"
	if queryErr := tx.Save(&contact).Error; queryErr != nil {
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	chat, chatErr := model.CreateChat(tx)
	if chatErr != nil {
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	userIds := []uint{contact.SenderId, contact.ReceiverId}
	chatUserErr := model.SaveChatUserBatch(chat.ID, userIds, tx)
	if chatUserErr != nil {
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	var receiverContactWithUser model.ContactWithUserAndChat
	receiverQueryErr := tx.Table("contacts AS c").
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
		`, authId).
		Joins(`
			JOIN users u
			ON u.id = c.sender_id
		`).
		Where("c.status = 'accepted'").
		Where("c.id = ?", contact.Id).
		First(&receiverContactWithUser).Error
	if receiverQueryErr != nil {
		tx.Rollback()
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	var senderContactWithUser model.ContactWithUserAndChat
	senderQueryErr := tx.Table("contacts AS c").
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
		`, contact.SenderId).
		Joins(`
			JOIN users u
			ON u.id = c.receiver_id
		`).
		Where("c.status = 'accepted'").
		Where("c.id = ?", contact.Id).
		First(&senderContactWithUser).Error
	if senderQueryErr != nil {
		tx.Rollback()
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	msg := map[string]interface{}{
		"contact": senderContactWithUser,
		"chat": chat,
		"type": "InviteAccepted",
		"target_id": contact.SenderId,
	}
	msgJson, parseErr := json.Marshal(msg)
	if parseErr != nil {
		tx.Rollback()
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	userKey := strconv.FormatUint(uint64(contact.SenderId), 10)
	redisCtx := context.Background()
	nodeHash, redisErr := c.redis.Get(redisCtx, userKey).Result()
	if redisErr == nil {
		rabbitErr := c.rabbit.Publish(
			"",
			nodeHash,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType: "application/json",
				Body: msgJson,
			},
		)
		if rabbitErr != nil {
			tx.Rollback()
			ctx.JSON(500, gin.H{"message": "Internal error"})
			return
		}
	}

	commitErr := tx.Commit().Error
	if commitErr != nil {
		tx.Rollback()
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	ctx.JSON(201, gin.H{
		"chat": chat,
		"contact": receiverContactWithUser,
	})
}
