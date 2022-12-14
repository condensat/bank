module git.condensat.tech/bank

go 1.13

require (
	code.condensat.tech/bank/secureid v0.1.0
	github.com/bsm/redislock v0.5.0
	github.com/emef/bitfield v0.0.0-20170503144143-7d3f8f823065
	github.com/go-redis/redis/v7 v7.2.0
	github.com/go-redis/redis_rate/v8 v8.0.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/rpc v1.2.0
	github.com/gorilla/sessions v1.2.0
	github.com/jinzhu/gorm v1.9.12
	github.com/joho/godotenv v1.3.0
	github.com/markbates/goth v1.64.0
	github.com/nats-io/nats-server/v2 v2.1.2 // indirect
	github.com/nats-io/nats.go v1.9.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/pquerna/otp v1.3.0
	github.com/rs/cors v1.7.0
	github.com/shengdoushi/base58 v1.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/thoas/stats v0.0.0-20190407194641-965cb2de1678
	github.com/urfave/negroni v1.0.0
	github.com/ybbus/jsonrpc v2.1.2+incompatible
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59
	golang.org/x/net v0.0.0-20190923162816-aa69164e4478
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
)

replace github.com/btcsuite/btcd => github.com/condensat/btcd v0.20.1-beta.0.20200424100000-5dc523e373e2

replace golang.org/x/crypto/openpgp => github.com/keybase/go-crypto/openpgp v0.0.0-20200123153347-de78d2cb44f4
