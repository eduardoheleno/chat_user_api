package controller

import (
	"nossochat_api/models"
	"nossochat_api/utils"

	"jwt_auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController interface {
	Store(ctx *gin.Context)
	Login(ctx *gin.Context)
	Search(ctx *gin.Context)
}

type userController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) UserController {
	return &userController{db}
}

func (c *userController) Store(ctx *gin.Context) {
	var user model.User
	err := ctx.ShouldBindJSON(&user)
	if err != nil {
		ctx.JSON(400, gin.H{"message": "Bad request"})
		return
	}

	isLoginRequest := false
	if emailErr := user.ValidateEmail(c.db, isLoginRequest); emailErr != nil {
		ctx.JSON(
			emailErr.(*model.EmailError).StatusCode,
			gin.H{"message": emailErr.Error()},
		)
		return
	}
	if passwordErr := user.ValidatePassword(); passwordErr != nil {
		ctx.JSON(400, gin.H{"message": passwordErr.Error()})
		return
	}
	if publicKeyErr := user.ValidatePublicKey(); publicKeyErr != nil {
		ctx.JSON(400, gin.H{"message": publicKeyErr.Error()})
		return
	}

	result := c.db.Create(&user)
	if result.Error != nil {
		ctx.JSON(500, gin.H{"message": result.Error.Error()})
		return
	}

	ctx.JSON(201, gin.H{"message": "User created"})
}

func (c *userController) Login(ctx *gin.Context) {
	var user model.User
	err := ctx.ShouldBindJSON(&user)
	if err != nil {
		ctx.JSON(400, gin.H{"message": "Bad request"})
		return
	}

	isLoginRequest := true;
	if emailErr := user.ValidateEmail(c.db, isLoginRequest); emailErr != nil {
		ctx.JSON(400, gin.H{"message": emailErr.Error()})
		return
	}
	if passwordErr := user.ValidatePassword(); passwordErr != nil {
		ctx.JSON(400, gin.H{"message": passwordErr.Error()})
		return
	}

	var dbUser model.User
	result := c.db.Where("email = ?", user.Email).First(&dbUser);
	if result.RowsAffected <= 0 {
		ctx.JSON(400, gin.H{"message": "User doesn't exists"})
		return
	}

	hashErr := util.ComparePassword(dbUser.Password, user.Password)
	if hashErr != nil {
		ctx.JSON(400, gin.H{"message": hashErr.Error()})
		return
	}

	jwtToken, jwtErr := jwt_auth.GenerateJwtToken(dbUser.ID, dbUser.Email)
	if jwtErr != nil {
		ctx.JSON(500, gin.H{"message": jwtErr.Error()})
		return
	}

	contacts, contactsErr := model.GetContactsByUser(c.db, dbUser.ID)
	if contactsErr != nil {
		ctx.JSON(500, gin.H{"message": "Couldn't retrieve contacts"})
		return
	}

	pendingSentInvites := []map[string]interface{}{}
	if queryErr := c.db.Table("contacts").
		Select(`
			contacts.*,
			u1.email as sender_email,
			u2.email as receiver_email
		`).
		Joins("right join users u1 on u1.id = contacts.sender_id").
		Joins("right join users u2 on u2.id = contacts.receiver_id").
		Where("contacts.sender_id = ?", dbUser.ID).
		Where("contacts.status = 'pending'").
		Find(&pendingSentInvites).Error; queryErr != nil {
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	pendingReceivedInvites := []map[string]interface{}{}
	if queryErr := c.db.Table("contacts").
		Select(`
			contacts.*,
			u1.email as sender_email,
			u2.email as receiver_email
		`).
		Joins("right join users u1 on u1.id = contacts.sender_id").
		Joins("right join users u2 on u2.id = contacts.receiver_id").
		Where("contacts.receiver_id = ?", dbUser.ID).
		Where("contacts.status = 'pending'").
		Find(&pendingReceivedInvites).Error; queryErr != nil {
		ctx.JSON(500, gin.H{"message": "Internal error"})
		return
	}

	ctx.JSON(200, gin.H{
		"token": jwtToken,
		"user_id": dbUser.ID,
		"contacts": contacts,
		"pending_sent_invites": pendingSentInvites,
		"pending_received_invites": pendingReceivedInvites,
	})
}

func (c *userController) Search(ctx *gin.Context) {
	searchParam := ctx.Query("email")
	userId, exists := ctx.Get("userId")

	if len(searchParam) <= 0 || !exists {
		ctx.JSON(400, gin.H{"message": "Wrong params"})
		return
	}

	var users []model.QueryUser
	err := c.db.
		Model(&model.User{}).
		Where("email LIKE ?", "%" + searchParam + "%").
		Where("id <> ?", userId).
		Limit(10).
		Find(&users).
		Error
	if err != nil {
		ctx.JSON(500, gin.H{"message": "Couldn't query to database"})
		return
	}

	ctx.JSON(200, users)
}
