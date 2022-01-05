module github.com/ONSdigital/dp-upload-service

go 1.16

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible

replace github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go v3.2.1-0.20210802184156-9742bd7fca1c+incompatible

require (
	github.com/ONSdigital/dp-component-test v0.6.3
	github.com/ONSdigital/dp-healthcheck v1.2.3
	github.com/ONSdigital/dp-net v1.2.0
	github.com/ONSdigital/dp-s3 v1.10.0 // indirect
	github.com/ONSdigital/dp-s3/v2 v2.0.0-beta.1
	github.com/ONSdigital/dp-vault v1.1.2
	github.com/ONSdigital/log.go/v2 v2.0.9
	github.com/aws/aws-sdk-go v1.42.29
	github.com/cucumber/godog v0.12.2
	github.com/gabriel-vasile/mimetype v1.4.0
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/maxcnunes/httpfake v1.2.4
	github.com/pkg/errors v0.9.1
	github.com/rdumont/assistdog v0.0.0-20201106100018-168b06230d14
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3 // indirect
)
