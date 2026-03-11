package validator

import (
	"encoding/json"
	"net/http"
)

type Validator interface {
	Validate() error
}

func DecodeAndValidate(r *http.Request, v Validator) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	defer r.Body.Close()
	return v.Validate()
}
