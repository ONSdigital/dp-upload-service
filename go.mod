module github.com/ONSdigital/dp-upload-service

go 1.16

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible
replace github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go v3.2.1-0.20210802184156-9742bd7fca1c+incompatible

require (
	github.com/ONSdigital/dp-component-test v0.6.3
	github.com/ONSdigital/dp-healthcheck v1.2.2
	github.com/ONSdigital/dp-net v1.2.0
	github.com/ONSdigital/dp-s3 v1.6.0
	github.com/ONSdigital/dp-vault v1.1.2
	github.com/ONSdigital/log.go/v2 v2.0.9
	github.com/aws/aws-sdk-go v1.29.9
	github.com/cucumber/godog v0.12.2
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.7.2
)
