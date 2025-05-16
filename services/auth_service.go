package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	db "github.com/dudeiebot/ad-ly/config"
	customizedError "github.com/dudeiebot/ad-ly/errors"
	"github.com/dudeiebot/ad-ly/helpers"
	"github.com/dudeiebot/ad-ly/mailer"
	"github.com/dudeiebot/ad-ly/models"
	"github.com/dudeiebot/ad-ly/queue"
	"github.com/dudeiebot/ad-ly/request"
	"github.com/dudeiebot/ad-ly/responses"
)

func RegisterUser(
	payload request.Register,
) (response responses.AuthResponse, err error, status int) {
	var user models.User
	_ = db.PostDb.Where("email = ?", payload.Email).Find(&user)

	if !user.Empty() {
		return response, customizedError.ErrEmaiAlreadyTaken, http.StatusNotAcceptable
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return responses.AuthResponse{}, helpers.ServerError(err), http.StatusInternalServerError
	}

	user = models.User{
		Id:        uuid.New().String(),
		Name:      payload.Name,
		Password:  string(hashedPassword),
		Email:     payload.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err = db.PostDb.Create(&user).Error; err != nil {
		return response, helpers.ServerError(err), http.StatusInternalServerError
	}

	token, err := helpers.GenerateAccessToken(context.Background(), user.Id)
	if err != nil {
		return response, helpers.ServerError(err), http.StatusInternalServerError
	}

	err = helpers.GenerateOtpToken(context.Background(), &user)
	if err != nil {
		return response, helpers.ServerError(err), http.StatusInternalServerError
	}

	return responses.AuthResponse{
		Token: token,
		User:  responses.GenerateUserResponse(user),
	}, nil, http.StatusOK
}

func VerifyUser(token string) (message map[string]string, err error, status int) {
	redisKey := "signup_otp_" + token
	var user models.User

	userId, err := db.Redis.Get(context.Background(), redisKey).Result()
	if err == redis.Nil {
		return nil, helpers.ServerError(err), http.StatusBadRequest
	} else if err != nil {
		return nil, helpers.ServerError(err), http.StatusInternalServerError
	}

	err = db.PostDb.Where("id = ?", userId).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helpers.ServerError(err), http.StatusNotFound
		}
		return nil, helpers.ServerError(err), http.StatusInternalServerError
	}
	err = db.PostDb.Model(&user).Update("email_verified_at", time.Now()).Error
	if err != nil {
		return nil, helpers.ServerError(err), http.StatusInternalServerError
	}

	err = db.Redis.Del(context.Background(), redisKey).Err()
	if err != nil {
		return nil, helpers.ServerError(err), http.StatusInternalServerError
	}
	return helpers.Message("email verified"), nil, http.StatusOK
}

func LoginUser(
	payload request.LoginUser,
) (response responses.AuthResponse, err error, status int) {
	var user models.User
	err = db.PostDb.Where("email = ?", payload.Email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response, customizedError.ErrInvalidCredentials, http.StatusUnauthorized
		}
		return response, helpers.ServerError(err), http.StatusInternalServerError
	}
	fmt.Println(payload.Email)

	if user.Empty() {
		return response, customizedError.ErrInvalidCredentials, http.StatusUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return response, customizedError.ErrInvalidCredentials, http.StatusUnauthorized
	}

	if !user.EmailVerified() {
		if err = helpers.CanSendVerification(context.Background(), user.Id); err != nil {
			return response, helpers.ServerError(err), http.StatusUnauthorized
		}
		err = helpers.GenerateOtpToken(context.Background(), &user)
		if err != nil {
			return response, helpers.ServerError(err), http.StatusInternalServerError
		}
		return response, customizedError.ErrEmailNotVerified, http.StatusUnauthorized
	}

	token, err := helpers.GenerateAccessToken(context.Background(), user.Id)
	if err != nil {
		return response, helpers.ServerError(err), http.StatusInternalServerError
	}

	return responses.AuthResponse{
		Token: token,
		User:  responses.GenerateUserResponse(user),
	}, nil, http.StatusOK
}

func ForgotPassword(
	payload request.ForgotPassword,
) (message map[string]string, err error, status int) {
	var user models.User
	_ = db.PostDb.Where("email = ?", payload.Email).First(&user).Error
	if err != nil {
		return nil, helpers.ServerError(err), http.StatusBadRequest
	}

	if !user.Empty() {
		token := uuid.New().String()
		err = db.Redis.Set(context.Background(), "forgot_password_"+token, user.Id, time.Hour*1).
			Err()
		if err != nil {
			return nil, helpers.ServerError(err), http.StatusInternalServerError
		}
		apiHost := db.GetApiHost()

		err = mailer.EnqueueEmailTask(queue.Client, mailer.EmailPayload{
			TemplateName: "forget_password",
			To:           user.Email,
			Subject:      "Reset Your Password",
			Data: map[string]interface{}{
				"password_reset": fmt.Sprintf("%s/auth/password_reset?token=%s", apiHost, token),
				"Name":           user.Name,
			},
		})
		if err != nil {
			return nil, helpers.ServerError(err), http.StatusInternalServerError
		}
	}
	return helpers.Message("Check Your Email"), nil, http.StatusOK
}

func PostForgot(payload request.PostForgot) (message map[string]string, err error, status int) {
	redisKey := "forgot_password_" + payload.Token
	var user models.User

	userId, err := db.Redis.Get(context.Background(), redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("invalid token"), http.StatusNotAcceptable
		}
		return nil, helpers.ServerError(err), http.StatusInternalServerError
	}

	_ = db.PostDb.Where("id = ?", userId).First(&user).Error
	if err != nil {
		return nil, helpers.ServerError(err), http.StatusInternalServerError
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, helpers.ServerError(err), http.StatusInternalServerError
	}

	err = db.PostDb.Model(&user).Updates(map[string]interface{}{"password": hashedPassword}).Error
	if err != nil {
		return nil, helpers.ServerError(err), http.StatusInternalServerError
	}

	err = db.Redis.Del(context.Background(), redisKey).Err()
	if err != nil {
		return nil, helpers.ServerError(err), http.StatusInternalServerError
	}

	return helpers.Message("Password Reset Completed"), nil, http.StatusOK
}
