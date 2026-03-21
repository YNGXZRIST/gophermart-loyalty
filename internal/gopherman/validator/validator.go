// Package validator decodes and validates request payloads.
package validator

import (
	"encoding/json"
	"net/http"
)

// Validator is implemented by request payloads that can self-validate.
type Validator interface {
	Validate() error
}

// DecodeAndValidate decodes JSON body into value and validates it.
func DecodeAndValidate(r *http.Request, v Validator) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	defer r.Body.Close()
	return v.Validate()
}
