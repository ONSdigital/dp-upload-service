module github.com/ONSdigital/dp-upload-service

go 1.19

replace (
	github.com/go-ldap/ldap/v3 => github.com/go-ldap/ldap/v3 v3.4.3
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.12.2
	github.com/spf13/cobra => github.com/spf13/cobra v1.4.0
)

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.187.0
	github.com/ONSdigital/dp-component-test v0.7.0
	github.com/ONSdigital/dp-healthcheck v1.5.0
	github.com/ONSdigital/dp-net/v2 v2.6.0
	github.com/ONSdigital/dp-s3/v2 v2.0.0-beta.2
	github.com/ONSdigital/dp-vault v1.1.2
	github.com/ONSdigital/log.go/v2 v2.3.0
	github.com/aws/aws-sdk-go v1.44.76
	github.com/cucumber/godog v0.12.2
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/rdumont/assistdog v0.0.0-20201106100018-168b06230d14
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.8.0
)

require (
	github.com/ONSdigital/dp-api-clients-go v1.43.0 // indirect
	github.com/ONSdigital/dp-mongodb-in-memory v1.2.0 // indirect
	github.com/ONSdigital/dp-net v1.5.0 // indirect
	github.com/chromedp/cdproto v0.0.0-20211126220118-81fa0469ad77 // indirect
	github.com/chromedp/chromedp v0.7.6 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/cucumber/gherkin-go/v19 v19.0.3 // indirect
	github.com/cucumber/messages-go/v16 v16.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.1.0 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-hclog v0.12.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.0 // indirect
	github.com/hashicorp/go-memdb v1.3.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.2 // indirect
	github.com/hashicorp/go-rootcerts v1.0.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/vault/api v1.0.5-0.20200527182800-ad90e0b39d2f // indirect
	github.com/hashicorp/vault/sdk v0.1.14-0.20200514144402-4bfac290c352 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/maxcnunes/httpfake v1.2.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/smartystreets/assertions v1.13.0 // indirect
	github.com/spf13/afero v1.8.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20220331220935-ae2d96664a29 // indirect
	golang.org/x/net v0.0.0-20220812174116-3211cb980234 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/text v0.3.8 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
