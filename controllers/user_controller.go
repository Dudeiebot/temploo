package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/dudeiebot/ad-ly/helpers"
	"github.com/dudeiebot/ad-ly/services"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	resp, err, status := services.GetUser(r)

	if err != nil {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(helpers.Message(err.Error()))
		return
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
	return
}
