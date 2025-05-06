package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dudeiebot/ad-ly/config"
	"github.com/dudeiebot/ad-ly/helpers"
	"github.com/dudeiebot/ad-ly/models"
)

type userCtxKey string

const (
	userKey userCtxKey = "user"
)

func AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", -1)

		unauthorized := helpers.Message("Unauthorized")

		if token == "" {
			w.Header().Set("Content/Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(unauthorized)
			return
		}

		tempToken, _ := helpers.ParseAccessToken(token)
		var foundUser models.User

		if tempToken == "" {
			tempToken = ""
		}

		userId := config.Redis.Get(r.Context(), "user_auth_"+tempToken).Val()
		_ = config.PostDb.Where("id = ?", userId).First(&foundUser).Error

		if foundUser.Empty() {
			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(http.StatusUnauthorized)

			_ = json.NewEncoder(w).Encode(unauthorized)

			return
		}
		r = r.WithContext(context.WithValue(r.Context(), userKey, foundUser))
		next.ServeHTTP(w, r)
	})
}

func GetUser(ctx context.Context) models.User {
	return ctx.Value(userKey).(models.User)
}

func IsUser(ctx context.Context, user models.User) bool {
	return ctx.Value(userKey).(models.User).Id == user.Id
}
