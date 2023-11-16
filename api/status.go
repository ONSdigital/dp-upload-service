package api

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/ONSdigital/dp-upload-service/files"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"net/http"
)

func StatusHandler(store files.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		authHeaderValue := req.Header.Get(request.AuthHeaderKey)
		augmentedContext := context.WithValue(req.Context(), config.AuthContextKey, authHeaderValue)

		path := mux.Vars(req)["path"]
		status, err := store.Status(req.Context(), path)
		if err != nil {
			log.Error(augmentedContext, "error getting status", err)
			switch err {
			case files.ErrFilesAPINotFound:
				writeError(w, buildErrors(err, "NotFound"), http.StatusNotFound)
			case files.ErrS3Download:
				writeError(w, buildErrors(err, "S3Download"), http.StatusServiceUnavailable)
			default:
				writeError(w, buildErrors(err, "InternalError"), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(status); err != nil {
			log.Error(augmentedContext, "error encoding status response", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
