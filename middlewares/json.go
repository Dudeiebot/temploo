package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/dudeiebot/ad-ly/helpers"
)

func AcceptJson(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func ValidateJson(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(helpers.Message("Invalid Json Format"))

			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))

		var jsonTest interface{}
		if len(body) > 0 && json.Unmarshal(body, &jsonTest) != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(helpers.Message("Invalid Json Format"))

			return
		}
		next.ServeHTTP(w, r)
	})
}
