package api

import (
	"context"

	"github.com/gorilla/mux"
)

//API provides a struct to wrap the api around
type API struct {
	Router     *mux.Router
	vault      VaultClienter
	s3Uploaded S3Clienter
}

//Setup function sets up the api and returns an api
func Setup(ctx context.Context, vault VaultClienter, r *mux.Router, s3Uploaded S3Clienter) *API {
	api := &API{
		Router:     r,
		s3Uploaded: s3Uploaded,
		vault:      vault,
	}
	r.HandleFunc("/hello", HelloHandler()).Methods("GET")
	return api
}
