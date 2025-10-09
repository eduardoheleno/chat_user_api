package model

import (
	"nossochat_api/utils"
	"time"

	"errors"
	"fmt"
	"net/mail"
	"crypto/ecdh"

	"gorm.io/gorm"
)

type User struct {
	ID uint `gorm:"primarykey"`
	Email    string `json:"email" gorm:"unique;not null"`
	Password string `json:"password" gorm:"not null"`
	PublicKey []byte `json:"public_key" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type QueryUser struct {
	ID uint `json:"id"`
	Email    string `json:"email" gorm:"unique;not null"`
	PublicKey []byte `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EmailError struct {
	StatusCode int
	Err        error
}

func (e *EmailError) Error() string {
	return fmt.Sprintf("%s", e.Err)
}

func (u *User) BeforeCreate(ctx *gorm.DB) (err error) {
	hashedPassword, err := util.HashPassword(u.Password)
	if err != nil {
		return err
	}

	u.Password = hashedPassword
	return nil
}

func (u User) ValidateEmail(db *gorm.DB, isLoginRequest bool) error {
	if len(u.Email) <= 0 {
		return &EmailError {
			StatusCode: 400,
			Err:        errors.New("E-mail is required"),
		}
	}

	_, isValidErr := mail.ParseAddress(u.Email)
	if isValidErr != nil {
		return &EmailError{
			StatusCode: 400,
			Err:        errors.New("Invalid e-mail"),
		}
	}

	if !isLoginRequest {
		result := db.Where("email = ?", u.Email).First(&User{})
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &EmailError{
				StatusCode: 500,
				Err:        result.Error,
			}
		} else if result.RowsAffected > 0 {
			return &EmailError{
				StatusCode: 400,
				Err:        errors.New("E-mail already registered"),
			}
		}
	}

	return nil
}

func (u User) ValidatePublicKey() error {
	_, err := ecdh.X25519().NewPrivateKey(u.PublicKey)
	if err != nil {
		return errors.New("Invalid public key")
	}

	return nil
}

func (u User) ValidatePassword() error {
	if len(u.Password) <= 0 {
		return errors.New("Password is required")
	}

	return nil
}
