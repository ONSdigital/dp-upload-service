package upload

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	s3client "github.com/ONSdigital/dp-s3"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

// Resumable represents resumable js upload query pararmeters
type Resumable struct {
	ChunkNumber      int    `schema:"resumableChunkNumber"`
	TotalChunks      int    `schema:"resumableTotalChunks"`
	ChunkSize        int    `schema:"resumableChunkSize"`
	CurrentChunkSize int    `schema:"resumableCurrentChunkSize"`
	TotalSize        int    `schema:"resumableTotalSize"`
	Type             string `schema:"resumableType"`
	Identifier       string `schema:"resumableIdentifier"`
	FileName         string `schema:"resumableFilename"`
	RelativePath     string `schema:"resumableRelativePath"`
	AliasName        string `schema:"aliasName"`
}

// createS3Request creates a S3 UploadRequest struct from a Resumable struct
func (resum *Resumable) createS3Request() *s3client.UploadPartRequest {
	log.Event(nil, "calling function s3 request", log.Data{"resumeidentifier": resum.Identifier})
	return &s3client.UploadPartRequest{
		UploadKey:   resum.Identifier,
		Type:        resum.Type,
		ChunkNumber: int64(resum.ChunkNumber),
		TotalChunks: resum.TotalChunks,
		FileName:    resum.FileName,
	}
}

// Uploader represents the necessary configuration for uploading a file
type Uploader struct {
	s3Client    S3Clienter
	vaultClient VaultClienter
	vaultPath   string
	s3Region    string
	s3Bucket    string
}

// New returns a new Uploader from the provided clients and vault path
func New(s3 S3Clienter, vc VaultClienter, vaultPath, s3Region, s3Bucket string) *Uploader {
	return &Uploader{
		s3Client:    s3,
		vaultClient: vc,
		vaultPath:   vaultPath,
		s3Region:    s3Region,
		s3Bucket:    s3Bucket,
	}
}

// CheckUploaded checks to see if a chunk has been uploaded
func (u *Uploader) CheckUploaded(w http.ResponseWriter, req *http.Request) {

	if err := req.ParseForm(); err != nil {
		log.Event(req.Context(), "error parsing form", log.ERROR, log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resum := new(Resumable)

	if err := decoder.Decode(resum, req.Form); err != nil {
		log.Event(req.Context(), "error decoding form", log.ERROR, log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err := u.s3Client.CheckPartUploaded(req.Context(), resum.createS3Request())
	if err != nil {
		log.Event(req.Context(), "error returned from check part uploaded", log.ERROR, log.Error(err))
		w.WriteHeader(statusCodeFromError(err))
		return
	}

	log.Event(req.Context(), "uploaded file successfully", log.INFO, log.Data{"file-name": resum.FileName, "uid": resum.Identifier, "size": resum.TotalSize})
	w.WriteHeader(http.StatusOK)

}

// Upload handles the uploading of a file to AWS s3
func (u *Uploader) Upload(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		log.Event(req.Context(), "error parsing form", log.ERROR, log.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resum := new(Resumable)

	if err := decoder.Decode(resum, req.Form); err != nil {
		log.Event(req.Context(), "error decoding form", log.WARN, log.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	content, _, err := req.FormFile("file")
	if err != nil {
		log.Event(req.Context(), "error getting file from form", log.ERROR, log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer content.Close()

	payload, err := ioutil.ReadAll(content)
	if err != nil {
		log.Event(req.Context(), "error reading file", log.ERROR, log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if u.vaultClient == nil {
		// Perform upload without PSK
		if err := u.s3Client.UploadPart(req.Context(), resum.createS3Request(), payload); err != nil {
			w.WriteHeader(statusCodeFromError(err))
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	vaultKey := "key"
	vaultKeyPath := u.vaultPath + "/" + resum.Identifier

	// Get PSK from Vault. If the vault PSK is not found for this file, then create one and use it
	var psk []byte
	pskStr, err := u.vaultClient.ReadKey(vaultKeyPath, vaultKey)
	if err != nil {
		// Create PSK and write it to Vault
		psk = createPSK()
		if err := u.vaultClient.WriteKey(vaultKeyPath, vaultKey, hex.EncodeToString(psk)); err != nil {
			log.Event(req.Context(), "error writing key to vault", log.ERROR, log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		// Use existing PSK found in Vault
		psk, err = hex.DecodeString(pskStr)
		if err != nil {
			log.Event(req.Context(), "error decoding key", log.ERROR, log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Perform upload using vault PSK
	if err = u.s3Client.UploadPartWithPsk(req.Context(), resum.createS3Request(), payload, psk); err != nil {
		w.WriteHeader(statusCodeFromError(err))
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

// handleError decides the HTTP status according to the provided error
func statusCodeFromError(err error) int {
	switch err.(type) {
	case *s3client.ErrNotUploaded:
		return http.StatusNotFound
	case *s3client.ErrListParts:
		return http.StatusNotFound
	case *s3client.ErrChunkNumberNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
	// TODO I would suggest considering S3 client errors to be '502 BAD gateway'
}

// GetS3URL returns an S3 URL for a requested path, and the client's region and bucket name.
// Path corresponds to the S3 object key
func (u *Uploader) GetS3URL(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Query().Get(":id")

	// Generate URL from region, bucket and S3 key defined by query
	s3Url, err := s3client.NewURL(u.s3Region, u.s3Bucket, path)
	if err != nil {
		log.Event(req.Context(), "error generating S3 URL from bucket and path", log.ERROR, log.Error(err),
			log.Data{"bucket": u.s3Bucket, "region": u.s3Region, "path": path})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	url, err := s3Url.String(s3client.PathStyle)
	if err != nil {
		log.Event(req.Context(), "error getting path-style S3 URL", log.ERROR, log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := struct {
		URL string `json:"url"`
	}{
		URL: url,
	}

	b, err := json.Marshal(body)
	if err != nil {
		log.Event(req.Context(), "error marshalling json", log.ERROR, log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func createPSK() []byte {
	key := make([]byte, 16)
	rand.Read(key)

	return key
}
