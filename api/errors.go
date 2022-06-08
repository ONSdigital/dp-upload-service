package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
)

type jsonError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type jsonErrors struct {
	Error []jsonError `json:"errors"`
}

func writeError(w http.ResponseWriter, errs jsonErrors, httpCode int) {
	w.WriteHeader(httpCode)
	if err := json.NewEncoder(w).Encode(&errs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func buildErrors(err error, code string) jsonErrors {
	return jsonErrors{Error: []jsonError{{Description: err.Error(), Code: code}}}
}

func buildValidationErrors(validationErrs validator.ValidationErrors) jsonErrors {
	jsonErrs := jsonErrors{Error: []jsonError{}}

	for _, validationErr := range validationErrs {
		desc := fmt.Sprintf("%s %s", validationErr.Field(), validationErr.Tag())
		jsonErrs.Error = append(jsonErrs.Error, jsonError{Code: "ValidationError", Description: desc})
	}
	return jsonErrs
}
