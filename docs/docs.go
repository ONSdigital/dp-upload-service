// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "Open Government Licence v3.0",
            "url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "returns a health check",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "upload"
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/upload": {
            "get": {
                "description": "checks to see if a chunk has been uploaded",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "upload"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "name": "aliasName",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "chunkNumber",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "chunkSize",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "currentChunkSize",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "fileName",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "identifier",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "relativePath",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "totalChunks",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "totalSize",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "type",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "post": {
                "description": "handles the uploading of a file to AWS s3",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "upload"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "name": "aliasName",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "chunkNumber",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "chunkSize",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "currentChunkSize",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "fileName",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "identifier",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "relativePath",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "totalChunks",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "name": "totalSize",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "name": "type",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/upload/{id}": {
            "get": {
                "description": "returns an S3 URL for a requested path, and the client's region and bucket name",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "upload"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "S3 object key",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        }
    },
    "tags": [
        {
            "name": "private"
        }
    ]
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:25100",
	BasePath:         "",
	Schemes:          []string{"http"},
	Title:            "dp-upload-service",
	Description:      "Digital Publishing resumable file upload service that handles writing to S3. It updates images through the CMS.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
