package responses

import (
	"github.com/dudeiebot/ad-ly/helpers"
	"github.com/dudeiebot/ad-ly/models"
)

type UserResponse struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func GenerateUserResponse(user models.User) UserResponse {
	return UserResponse{
		Id:            user.Id,
		Name:          user.Name,
		Email:         user.Email,
		EmailVerified: user.EmailVerified(),
		CreatedAt:     helpers.JSONTime{Time: user.CreatedAt}.Json(),
		UpdatedAt:     helpers.JSONTime{Time: user.UpdatedAt}.Json(),
	}
}
