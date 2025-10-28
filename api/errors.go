package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
)

type JsonError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type JsonErrors struct {
	Error []JsonError `json:"errors"`
}

func writeError(w http.ResponseWriter, errs JsonErrors, httpCode int) {
	w.WriteHeader(httpCode)
	if err := json.NewEncoder(w).Encode(&errs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func buildErrors(err error, code string) JsonErrors {
	return JsonErrors{Error: []JsonError{{Description: err.Error(), Code: code}}}
}

func buildValidationErrors(validationErrs validator.ValidationErrors) JsonErrors {
	jsonErrs := JsonErrors{Error: []JsonError{}}

	for _, validationErr := range validationErrs {
		desc := fmt.Sprintf("%s %s", validationErr.Field(), validationErr.Tag())
		jsonErrs.Error = append(jsonErrs.Error, JsonError{Code: "ValidationError", Description: desc})
	}
	return jsonErrs
}
