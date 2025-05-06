package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/thedevsaddam/govalidator"

	"github.com/dudeiebot/ad-ly/helpers"
	"github.com/dudeiebot/ad-ly/request"
	"github.com/dudeiebot/ad-ly/services"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var req request.Register

	rules := govalidator.MapData{
		"name":     []string{"required", "alpha_space"},
		"email":    []string{"required", "email"},
		"password": []string{"regex:^.{8,}", "required"},
	}

	opt := govalidator.Options{
		Rules:   rules,
		Request: r,
		Data:    &req,
	}

	validationErrors := helpers.ValidateRequest(opt, "json")

	if len(validationErrors) != 0 {
		helpers.ReturnValidatorErrors(w, validationErrors)
		return
	}

	resp, err, status := services.RegisterUser(req)

	if err != nil {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(helpers.Message(err.Error()))
		return
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
	return
}

func VerifyUser(w http.ResponseWriter, r *http.Request) {
	var req request.VerifyUser
	req.Token = r.URL.Query().Get("token")

	rules := govalidator.MapData{
		"token": []string{"required", "alpha_num"},
	}

	opts := govalidator.Options{
		Rules:   rules,
		Request: r,
		Data:    &req,
	}

	validationErrors := helpers.ValidateRequest(opts, "query")

	if len(validationErrors) != 0 {
		helpers.ReturnValidatorErrors(w, validationErrors)
		return
	}

	resp, err, status := services.VerifyUser(req.Token)
	if err != nil {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(helpers.Message(err.Error()))
		return
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
	return
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req request.LoginUser

	rules := govalidator.MapData{
		"email":    []string{"required", "email"},
		"password": []string{"regex:^.{8,}", "required"},
	}

	opts := govalidator.Options{
		Rules:   rules,
		Request: r,
		Data:    &req,
	}

	validationErrors := helpers.ValidateRequest(opts, "json")

	if len(validationErrors) != 0 {
		helpers.ReturnValidatorErrors(w, validationErrors)
		return
	}

	resp, err, status := services.LoginUser(req)

	if err != nil {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(helpers.Message(err.Error()))
		return
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
	return
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req request.ForgotPassword

	rules := govalidator.MapData{
		"email": []string{"required", "email"},
	}

	opts := govalidator.Options{
		Rules:   rules,
		Request: r,
		Data:    &req,
	}

	validationErrors := helpers.ValidateRequest(opts, "json")
	if len(validationErrors) != 0 {
		helpers.ReturnValidatorErrors(w, validationErrors)
	}

	resp, err, status := services.ForgotPassword(req)

	if err != nil {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(helpers.Message(err.Error()))
		return
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
	return
}

func PostForgot(w http.ResponseWriter, r *http.Request) {
	var req request.PostForgot
	req.Token = r.URL.Query().Get("token")

	rules := govalidator.MapData{
		"password": []string{"regex:^.{8,}", "required"},
		"token":    []string{"required", "uuid"},
	}

	opts := govalidator.Options{
		Rules:   rules,
		Request: r,
		Data:    &req,
	}

	validationErrors := helpers.ValidateRequest(opts, "json")
	if len(validationErrors) != 0 {
		helpers.ReturnValidatorErrors(w, validationErrors)
	}

	resp, err, status := services.PostForgot(req)

	if err != nil {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(helpers.Message(err.Error()))
		return
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
	return
}
