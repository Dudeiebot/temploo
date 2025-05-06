package helpers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/thedevsaddam/govalidator"
)

func ValidateRequest(opts govalidator.Options, method string) url.Values {
	var e url.Values

	v := govalidator.New(opts)

	switch method {
	case "json":
		e = v.ValidateJSON()
		break
	case "struct":
		e = v.ValidateStruct()
		break
	case "query":
		e = v.Validate()
	}
	return e
}

func ReturnValidatorErrors(w http.ResponseWriter, e url.Values) {
	err := map[string]interface{}{"message": "This data entity are invalids", "errors": e}
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(err)
}
