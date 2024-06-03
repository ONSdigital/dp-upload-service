# DP Upload Service

## Introduction

The Upload Service is part of the [Static Files System](https://github.com/ONSdigital/dp-static-files-compose).
This service is responsible for storing the metadata and state of files.

The service enables consumers to upload a large file (Maximum 50GB) in multiple parts (Maximum 1000 - Each part must be
5MB, except the last part, which can be smaller)

The file will be stored in AWS S3.

It used the [Files API](https://github.com/ONSdigital/dp-files-api) to store the information about a file and is state.

### REST API

The service is fully documented in [Swagger Docs](swagger.yaml)

### Deprecated Endpoints

Upload service has been updated with a new upload endpoint `/upload-new`. This endpoint enables consumers to upload any
type of file and uses [Files API](https://github.com/ONSdigital/dp-files-api) to store the relevant metadata.

The old URL `/upload` requires the uploading consumer to store its own metadata about the file.

Once all existing consumers have been moved from using `/upload` to using `/upload-new` all the existing `/upload`
endpoints
and their related code can be removed.

### S3 Buckets

The `/upload` endpoint saves files to the S3 bucket configured in the `UPLOAD_BUCKET_NAME` environment
variable. Once the file
is uploaded, other services within DP take the file for further processing. Once the processing is complete the new files
are stored in a separate S3 bucket ready for download/decryption.

The `/upload-new` endpoint saves files to the S3 bucket configured in
the `STATIC_FILES_ENCRYPTED_BUCKET_NAME` environment
variable. This S3 bucket is the same one used for uploading files at the end of CMD Dataset file processing.

## Getting started

* Run `make docker-local`
* Run (inside container) `make debug`

## Dependencies

* No further dependencies other than those defined in `go.mod`

## Configuration

| Environment variable               | Default               | Description                                                                                                        |
|------------------------------------|-----------------------|--------------------------------------------------------------------------------------------------------------------|
| BIND_ADDR                          | :25100                | The host and port to bind to                                                                                       |
| AWS_REGION                         | eu-west-2             | S3 region to use. This region has to match the region where the bucket was created                                 |
| UPLOAD_BUCKET_NAME                 | testing               | Name of the S3 bucket that dataset uploads are sent to                                                             | 
| STATIC_FILES_ENCRYPTED_BUCKET_NAME | -                     | Name of the S3 bucket that static file uploads are sent to                                                         | 
| GRACEFUL_SHUTDOWN_TIMEOUT          | 5s                    | The graceful shutdown timeout in seconds (`time.Duration` format)                                                  |
| HEALTHCHECK_INTERVAL               | 30s                   | Time between self-healthchecks (`time.Duration` format)                                                            |
| HEALTHCHECK_CRITICAL_TIMEOUT       | 90s                   | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format) |
| FILES_API_URL                      | -                     |                                                                                                                    |
| LOCALSTACK_HOST                    | -                     | The hostname of the localstack server used for integration testing                                                 |

## Testing uploads and downloading using cURL

### Uploading

To upload a file, send a `POST`request in a URI, pass the query string parameters used by `Resumable struct` [here](upload/upload.go) and pass the file as form-data, for example:

```
curl -i -X POST -H 'content-type: multipart/form-data' -H 'X-Florence-Token: ' -F file=@test_results1.csv http://localhost:25100/upload-new?resumableFilename=countries-short.csv&resumableChunkNumber=1&resumableType=text/csv&resumableTotalChunks=1&resumableChunkSize=5242880&aliasName=countries-short.csv&resumableTotalSize=25222&resumableIdentifier=30052415109-countries-short-csv&licenceUrl=https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/&isPublishable=False&licence=Open-Government-Licence-v3.0&path=testingpath&collectionId=collectionId-fromflorencesandboxcollectionurl
```
The file can then be viewed as XML in a browser at http://localhost:25100/upload-new

### Downloading

To download an uploaded file, make a `GET` request in a URI that uses the bucket name endpoint `/testing` followed by the key `<key>path/filename</key>`, for example: 

```
curl -X GET -L -O http://localhost:14566/testing/testingpath/test_results1.csv 
```

The command downloads the uploaded file to the directory from which it is run.

## API Client

There is an [API Client](https://github.com/ONSdigital/dp-api-clients-go/tree/main/upload) for the Upload API this is
part
of [dp-api-clients-go](https://github.com/ONSdigital/dp-api-clients-go) package.

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2022, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
