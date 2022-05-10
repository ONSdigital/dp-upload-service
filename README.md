# DP Upload Service

## Introduction
 
The Upload Service is part of the [Static Files System](https://github.com/ONSdigital/dp-static-files-compose).
This service is responsible for storing the metadata and state of files. 

The service enables consumers to upload a large file (Maximum 50GB) in multiple parts (Maximum 1000 - Each part must be 
5MB appart from the last part the can be smaller) 

The file will be stored in AWS S3. Each file will be encrypted with a unique key ensuring that is a files key is cracked
only the data within one file is compromised.

It used the [Files API](https://github.com/ONSdigital/dp-files-api) to store the information about a file and is state. 

### REST API

The service is fully documented in [Swagger Docs](swagger.yaml)

### Depricated Endpoints

Upload service has been updated with a new upload endpoint `/upload-new`. This endpoint enables consumers to upload any
type of file and used [Files API](https://github.com/ONSdigital/dp-files-api) to store the relivant metadata.

The old URL `/upload` requires the uploading consumer to store its own metadata about the file making the download jouney
quite complex. 

Once all existing consumers have been moved from using `/upload` to using `/upload-new` all the existing `/upload` endpoints
and their related code can be removed.

## Getting started

* Run `make docker-local`
* Run (inside container) `make debug`

Please note that encryption is enabled by default. To disable it set `ENCRYPTION_DISABLED=true`. If you wish to run with 
encryption enabled, encryption is always enabled for the `/upload-new` endpoint

## Dependencies

* No further dependencies other than those defined in `go.mod`

## Configuration

| Environment variable         | Default                           | Description                                                                                                          |
|------------------------------|-----------------------------------|----------------------------------------------------------------------------------------------------------------------|
| BIND_ADDR                    | :25100                            | The host and port to bind to                                                                                         |
| AWS_REGION                   | eu-west-1                         | S3 region to use. This region has to match the region where the bucket was created                                   |
| UPLOAD_BUCKET_NAME           | dp-frontend-florence-file-uploads | Name of the S3 bucket that dataset uploads are sent to                                                               | 
| ENCRYPTION_DISABLED          | false                             | Determines whether encryption is disabled or enabled                                                                 |    
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                                | The graceful shutdown timeout in seconds (`time.Duration` format)                                                    |
| HEALTHCHECK_INTERVAL         | 30s                               | Time between self-healthchecks (`time.Duration` format)                                                              |
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                               | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)   |
| VAULT_TOKEN                  | -                                 | Vault token required for the client to talk to vault. (Use `make debug` to create a vault token)                     |
| VAULT_ADDR                   | http://localhost:8200             | The vault address                                                                                                    |
| VAULT_PATH                   | secret/shared/psk                 | The path where the psks will be stored in vault                                                                      |
| FILES_API_URL                | -                                 |                                                                                                                      |

## To Test using Curl

To test upload functionality using `curl`, you need to pass the following query string parameters in the URL -  to satisfy the schema mentioned in the `Resumable` struct and pass the file as form-data

Please refer [Resumable struct](upload/upload.go).

* `curl -i -X POST -H 'content-type: multipart/form-data' -F file=@README.md 'http://localhost:25100/upload\?resumableFilename=README.md&resumableChunkNumber=1&resumableType=text/plain&resumableTotalChunks=1&resumableIdentifier=<KEY_MATCHING_VAULT_SECRET_KEY>&resumableChunkSize=1000000&aliasName=somealias'`

## API Client

There is an [API Client](https://github.com/ONSdigital/dp-api-clients-go/tree/main/upload) for the Upload API this is part
of [dp-api-clients-go](https://github.com/ONSdigital/dp-api-clients-go) package.

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
