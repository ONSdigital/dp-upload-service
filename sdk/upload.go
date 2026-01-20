package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/ONSdigital/dp-upload-service/api"
)

const (
	chunkSize   = 5 * 1024 * 1024
	maxChunks   = 10000
	maxFileSize = chunkSize * maxChunks
)

type ChunkInfo struct {
	Current int
	Total   int
}

// Upload uploads a file in chunks to the upload service via the /upload-new endpoint with the provided metadata and headers
func (cli *Client) Upload(ctx context.Context, fileContent io.ReadCloser, metadata api.Metadata, headers Headers) error {
	if metadata.SizeInBytes > maxFileSize {
		return ErrFileTooLarge
	}

	totalChunks := (metadata.SizeInBytes + chunkSize - 1) / chunkSize
	chunkInfo := ChunkInfo{Total: totalChunks}

	for i := 1; i <= totalChunks; i++ {
		chunkInfo.Current = i

		reqBody, contentType, err := createUploadRequestBody(chunkInfo, fileContent, metadata)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/upload-new", cli.hcCli.URL), reqBody)
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", contentType)
		headers.Add(req)

		resp, err := cli.hcCli.Client.Do(ctx, req)
		if err != nil {
			closeResponseBody(ctx, resp)
			return err
		}

		statusCode := resp.StatusCode
		if statusCode != http.StatusOK && statusCode != http.StatusCreated {
			jsonErrors, err := unmarshalJsonErrors(resp.Body)
			if err != nil {
				return err
			}
			closeResponseBody(ctx, resp)
			return &APIError{
				StatusCode: statusCode,
				Errors:     jsonErrors,
			}
		}

		closeResponseBody(ctx, resp)
	}

	return nil
}

// createUploadRequestBody creates a multipart/form-data request body for the given chunk and metadata
func createUploadRequestBody(chunkInfo ChunkInfo, fileContent io.ReadCloser, metadata api.Metadata) (*bytes.Buffer, string, error) {
	reqBuff := &bytes.Buffer{}
	formWriter := multipart.NewWriter(reqBuff)

	contentChunk, contentChunkLength, err := chunkReader(fileContent)
	if err != nil {
		return nil, "", err
	}

	err = writeMetadataFormFields(formWriter, metadata, chunkInfo)
	if err != nil {
		return nil, "", err
	}

	_, err = writeFileFormField(formWriter, contentChunk, contentChunkLength, filepath.Base(metadata.Path))
	if err != nil {
		return nil, "", err
	}

	if err := formWriter.Close(); err != nil {
		return nil, "", err
	}

	return reqBuff, formWriter.FormDataContentType(), nil
}

// chunkReader reads a chunk of data from the provided fileContent ReadCloser
func chunkReader(fileContent io.ReadCloser) (io.Reader, int, error) {
	readBuff := make([]byte, chunkSize)
	bytesRead, err := io.ReadFull(fileContent, readBuff)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, 0, err
	}

	var outBuff []byte
	if bytesRead == chunkSize {
		outBuff = readBuff
	} else if bytesRead < chunkSize {
		outBuff = readBuff[:bytesRead]
	}

	return bytes.NewReader(outBuff), len(outBuff), nil
}

// writeMetadataFormFields writes the metadata fields as form fields in the multipart form
func writeMetadataFormFields(formWriter *multipart.Writer, metadata api.Metadata, chunkInfo ChunkInfo) error {
	if metadata.CollectionId != nil {
		if err := formWriter.WriteField("collectionId", *metadata.CollectionId); err != nil {
			return err
		}
	}

	if metadata.BundleId != nil {
		if err := formWriter.WriteField("bundleId", *metadata.BundleId); err != nil {
			return err
		}
	}

	formFields := map[string]string{
		"path":                 metadata.Path,
		"isPublishable":        strconv.FormatBool(*metadata.IsPublishable),
		"title":                metadata.Title,
		"licence":              metadata.Licence,
		"licenceUrl":           metadata.LicenceUrl,
		"resumableType":        metadata.Type,
		"resumableTotalSize":   fmt.Sprintf("%d", metadata.SizeInBytes),
		"resumableChunkNumber": fmt.Sprintf("%d", chunkInfo.Current),
		"resumableTotalChunks": fmt.Sprintf("%d", chunkInfo.Total),
		"resumableFilename":    filepath.Base(metadata.Path),
	}

	if metadata.DatasetID != "" {
		formFields["datasetId"] = metadata.DatasetID
	}
	if metadata.Edition != "" {
		formFields["edition"] = metadata.Edition
	}
	if metadata.Version != "" {
		formFields["version"] = metadata.Version
	}

	for field, value := range formFields {
		if err := formWriter.WriteField(field, value); err != nil {
			return err
		}
	}

	return nil
}

// writeFileFormField writes the file content as a form file field in the multipart form
func writeFileFormField(formWriter *multipart.Writer, fileContent io.Reader, fileSizeInBytes int, filename string) (int, error) {
	part, err := formWriter.CreateFormFile("file", filename)
	if err != nil {
		return 0, err
	}

	fileContentBytes := make([]byte, fileSizeInBytes)

	_, err = fileContent.Read(fileContentBytes)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return 0, err
	}

	return part.Write(fileContentBytes)
}
