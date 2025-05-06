package helpers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dudeiebot/ad-ly/config"
	"github.com/dudeiebot/ad-ly/errors"
	"github.com/dudeiebot/ad-ly/mailer"
	"github.com/dudeiebot/ad-ly/models"
	"github.com/dudeiebot/ad-ly/queue"
)

func generateAlphaNumericToken(length int) (string, error) {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	token := make([]byte, length)

	for i := range token {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		token[i] = charset[num.Int64()]
	}

	return string(token), nil
}

func GenerateOtpToken(ctx context.Context, user *models.User) error {
	otpToken, err := generateAlphaNumericToken(10)
	if err != nil {
		return err
	}

	redisKey := "signup_otp_" + otpToken

	err = config.Redis.Set(ctx, redisKey, user.Id, time.Minute*10).Err()
	if err != nil {
		return err
	}

	apiHost := config.GetApiHost()

	err = mailer.EnqueueEmailTask(queue.Client, mailer.EmailPayload{
		TemplateName: "signup_otp",
		To:           user.Email,
		Subject:      "Verify Your Email",
		Data: map[string]interface{}{
			"verification_link": fmt.Sprintf("%s/auth/verify-email?token=%s", apiHost, otpToken),
			"Name":              user.Name,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func CanSendVerification(ctx context.Context, userId string) error {
	cooldownKey := "verify_cooldown_" + userId

	_, err := config.Redis.Get(ctx, cooldownKey).Result()
	if err == nil {
		return errors.ErrCantSendVerificationMail
	} else if err != redis.Nil {
		return err
	}

	err = config.Redis.Set(ctx, cooldownKey, true, 10*time.Minute).Err()
	if err != nil {
		return err
	}

	return nil
}
