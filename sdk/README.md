# dp-upload-service SDK

## Overview

This SDK provides a client for interacting with the dp-upload-service. It is intended to be consumed by services that require endpoints from the dp-upload-service. It also provides healthcheck functionality, mocks and structs for easy integration, testing and error handling.

## Available client methods

| Name | Description |
|------|-------------|
| [`URL`](#url) | Returns the URL used by the Client |
| [`Health`](#health) | Returns the `health.Client` used by the Client |
| [`Checker`](#checker) | Calls the `health.Client`'s `Checker` method |
| [`Upload`](#upload) | Uploads a file in chunks to the upload service via the `/upload-new` endpoint with the provided metadata and headers |

## Instantiation

Example using `New`:

```go
package main

import "github.com/ONSdigital/dp-upload-service/sdk"

func main() {
    client := sdk.New("http://localhost:25100")
}
```

Example using `NewWithHealthClient`:

```go
import (
    "github.com/ONSdigital/dp-api-clients-go/v2/health"
    "github.com/ONSdigital/dp-upload-service/sdk"
)

func main() {
    existingHealthClient := health.NewClient("existing-service-name", "http://localhost:8080")

    client := sdk.NewWithHealthClient(existingHealthClient)
}
```

## Example usage of client

This example demonstrates how the `Upload()` function could be used:

```go
package main

import (
    "context"
    "os"

    "github.com/ONSdigital/dp-upload-service/api"
    "github.com/ONSdigital/dp-upload-service/sdk"
)

func main() {
    client := sdk.New("http://localhost:25100")

    // Prepare file content
    fileContent, err := os.Open("path/to/file.csv")
    if err != nil {
        panic(err)
    }
    defer fileContent.Close()

    // Prepare metadata
    isPublishable := true
    collectionID := "example-collection-id"

    // Build Metadata struct
    metadata := api.Metadata{
        Path:          "path/to/file.csv",
        IsPublishable: &isPublishable,
        CollectionId:  &collectionID, // Only one of BundleId or CollectionId should be set
        Title:         "Example Title",
        SizeInBytes:   123, // Replace with actual file size
        Type:          "text/csv",
        Licence:       "Example Licence",
        LicenceUrl:    "http://example.com/licence",
    }

    // Prepare Headers for authentication
    headers := sdk.Headers{
        ServiceAuthToken: "example-auth-token",
    }

    // Call Upload method
    err = client.Upload(context.Background(), fileContent, metadata, headers)
    if err != nil {
        // Distinguish between API errors and other errors
        apiErr, ok := err.(*sdk.APIError)
        if ok {
            // Type is *sdk.APIError so we can access all the following fields:
            // apiErr.StatusCode
            // apiErr.Errors // This is an array that can be looped through
            // apiErr.Errors.Error[0].Code
            // apiErr.Errors.Error[0].Description
            // apiErr.Error()
        } else {
            // Handle non-API errors
        }
    }
}
```

## Available Functionality

### Checker

```go
import "github.com/ONSdigital/dp-healthcheck/healthcheck"

check := &healthcheck.CheckState{}
err := client.Checker(ctx, check)
```

### Health

```go
healthClient := client.Health()
```

### URL

```go
url := client.URL()
```

### Upload

```go
import (
    "os"

    "github.com/ONSdigital/dp-upload-service/api"
    "github.com/ONSdigital/dp-upload-service/sdk"
)

// Prepare file content
fileContent, err := os.Open("path/to/file.csv")
if err != nil {
    panic(err)
}
defer fileContent.Close()

// Prepare metadata
isPublishable := true
collectionID := "example-collection-id" // Only one of BundleId or CollectionId should be set

// Build Metadata struct
metadata := api.Metadata{
    Path:          "path/to/file.csv",
    IsPublishable: &isPublishable,
    CollectionId:  &collectionID,
    Title:         "Example Title",
    SizeInBytes:   123, // Replace with actual file size
    Type:          "text/csv",
    Licence:       "Example Licence",
    LicenceUrl:    "http://example.com/licence",
}

// Prepare Headers for authentication
headers := sdk.Headers{
    ServiceAuthToken: "example-auth-token",
}

// Call Upload method
err = client.Upload(context.Background(), fileContent, metadata, headers)
```

## Additional Information

### Errors

The [`APIError`](errors.go) struct allows the user to distinguish if an error is a generic error or an API error, therefore allowing access to more detailed fields. This is shown in the [Example usage of client](#example-usage-of-client) section.

### Headers

The [`Headers`](headers.go) struct allows the user to provide an Authorization header if required. This is shown in the [Example usage of client](#example-usage-of-client) section. The `"Bearer "` prefix will be added automatically.

### Mocks

To simplify testing, all functions provided by the client have been defined in the [`Clienter` interface](interface.go). This allows the user to use [auto-generated mocks](mocks/) within unit tests.

Example of how to define a mock clienter:

```go
import (
    "context"
    "io"
    "testing"

    "github.com/ONSdigital/dp-upload-service/api"
    "github.com/ONSdigital/dp-upload-service/sdk"
    "github.com/ONSdigital/dp-upload-service/sdk/mocks"
)

func Test(t *testing.T) {
    mockClient := mocks.ClienterMock{
        UploadFunc: func(ctx context.Context, fileContent io.ReadCloser, metadata api.Metadata, headers sdk.Headers) error {
            // Setup mock behaviour here
            return nil
        },
        // Other methods can be mocked if needed
    }
}
```
