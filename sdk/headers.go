package sdk

import (
	"net/http"

	dprequest "github.com/ONSdigital/dp-net/v3/request"
)

type Headers struct {
	ServiceAuthToken string
}

// Add adds any non-empty headers to the provided http.Request
func (h *Headers) Add(req *http.Request) {
	if h.ServiceAuthToken != "" {
		dprequest.AddServiceTokenHeader(req, h.ServiceAuthToken)
	}
}
