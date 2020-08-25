dp-upload-service
================
Digital publishing resemable file upload service that handles on the fly encryption and writing to S3. It updates images through the CMS

### Getting started

* Run `make debug`

### Dependencies

* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable         | Default                           | Description
| ---------------------------- | ---------                         | -----------
| BIND_ADDR                    | :25100                            | The host and port to bind to
| AWS_REGION                   | eu-west-1                         | S3 region to use. This region has to match the region where the bucket was created
| UPLOAD_BUCKET_NAME           | dp-frontend-florence-file-uploads | Name of the S3 bucket that dataset uploads are sent to 
| ENCRYPTION_DISABLED          | false                             | Determines whether encryption is disabled or enabled    
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                                | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL         | 30s                               | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                               | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| VAULT_TOKEN                  | -                                 | Vault token required for the client to talk to vault. (Use `make debug` to create a vault token)
| VAULT_ADDR                   | http://localhost:8200             | The vault address
| VAULT_PATH                   | secret/shared/psk                 | The path where the psks will be stored in vault

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2020, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

