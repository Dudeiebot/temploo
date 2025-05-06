package services

import (
	"net/http"

	"github.com/dudeiebot/ad-ly/config"
	"github.com/dudeiebot/ad-ly/helpers"
	"github.com/dudeiebot/ad-ly/middlewares"
	"github.com/dudeiebot/ad-ly/models"
	"github.com/dudeiebot/ad-ly/responses"
)

func GetUser(r *http.Request) (response responses.UserResponse, err error, status int) {
	var user models.User
	userId := middlewares.GetUser(r.Context()).Id

	err = config.PostDb.Where("id = ?", userId).Find(&user).Error
	if err != nil {
		return response, helpers.ServerError(err), http.StatusInternalServerError
	}

	return responses.GenerateUserResponse(user), nil, http.StatusOK
}
