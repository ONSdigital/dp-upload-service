module github.com/ONSdigital/dp-upload-service

go 1.16

replace (
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go v3.2.1-0.20210802184156-9742bd7fca1c+incompatible
	github.com/go-ldap/ldap/v3 => github.com/go-ldap/ldap/v3 v3.4.3
	github.com/pkg/sftp => github.com/pkg/sftp v1.13.4
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.12.2
	github.com/spf13/cobra => github.com/spf13/cobra v1.4.0
)

require (
	github.com/ONSdigital/dp-component-test v0.6.3
	github.com/ONSdigital/dp-healthcheck v1.2.3
	github.com/ONSdigital/dp-net/v2 v2.1.0
	github.com/ONSdigital/dp-s3/v2 v2.0.0-beta.2
	github.com/ONSdigital/dp-vault v1.1.2
	github.com/ONSdigital/log.go/v2 v2.1.0
	github.com/aws/aws-sdk-go v1.42.47
	github.com/cucumber/godog v0.12.2
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
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
)
