package api

import (
	"context"

	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

//API provides a struct to wrap the api around
type API struct {
	Router     *mux.Router
	s3Uploaded S3Clienter
}

//Setup function sets up the api and returns an api
func Setup(
	ctx context.Context,
	r *mux.Router,
	s3Uploaded S3Clienter,
) *API {
	api := &API{
		Router:     r,
		s3Uploaded: s3Uploaded,
	}

	r.HandleFunc("/hello", HelloHandler()).Methods("GET")
	return api
}

//Close fuction shuts down dependenices
func (*API) Close(ctx context.Context) error {
	// Close any dependencies
	log.Event(ctx, "graceful shutdown of api complete", log.INFO)
	return nil
}
