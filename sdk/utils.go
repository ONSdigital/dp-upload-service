package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ONSdigital/dp-upload-service/api"
	"github.com/ONSdigital/log.go/v2/log"
)

// closeResponseBody closes the response body and logs an error if it fails
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "failed to close response body", err)
		}
	}
}

// unmarshalJsonErrors unmarshals the JSON errors from the response body.
// This function assumes the response body JSON structure matches api.JsonErrors
func unmarshalJsonErrors(body io.ReadCloser) (*api.JsonErrors, error) {
	if body == nil {
		return nil, nil
	}

	var jsonErrors api.JsonErrors

	bytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &jsonErrors); err != nil {
		return nil, err
	}

	return &jsonErrors, nil
}
