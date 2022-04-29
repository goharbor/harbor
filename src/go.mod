module github.com/goharbor/harbor/src

go 1.17

replace github.com/goharbor/harbor => ../

require (
	github.com/Masterminds/semver v1.4.2
	github.com/aliyun/alibaba-cloud-sdk-go v0.0.0-20190726115642-cd293c93fd97
	github.com/aws/aws-sdk-go v1.19.47
	github.com/beego/beego v1.12.8
	github.com/beego/i18n v0.0.0-20140604031826-e87155e8f0c0
	github.com/bmatcuk/doublestar v1.1.1
	github.com/casbin/casbin v1.7.0
	github.com/coreos/go-oidc v2.0.0+incompatible
	github.com/dghubble/sling v1.1.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/garyburd/redigo v1.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gocraft/work v0.5.1
	github.com/golang-migrate/migrate v3.3.0+incompatible
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.6.2
	github.com/graph-gophers/dataloader v5.0.0+incompatible
	github.com/justinas/alice v0.0.0-20171023064455-03f45bd4b7da
	github.com/lib/pq v1.1.0
	github.com/olekukonko/tablewriter v0.0.1
	github.com/opencontainers/go-digest v1.0.0-rc0
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.0.0
	github.com/stretchr/testify v1.4.0
	github.com/theupdateframework/notary v0.6.1
	golang.org/x/crypto v0.0.0-20220427172511-eb4f295cb31f
	golang.org/x/oauth2 v0.0.0-20220223155221-ee480838109b
	gopkg.in/ldap.v2 v2.5.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apimachinery v0.0.0-20180704011316-f534d624797b
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/helm v2.9.1+incompatible
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/Knetic/govaluate v3.0.0+incompatible // indirect
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/Unknwon/goconfig v0.0.0-20160216183935-5f601ca6ef4d // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bugsnag/bugsnag-go v1.5.2 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudflare/cfssl v0.0.0-20190510060611-9c027c93ba9e // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go v0.0.0-20160303222718-d30aec9fd63c // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/certificate-transparency-go v1.0.21 // indirect
	github.com/google/go-querystring v0.0.0-20170111101155-53e6ce116135 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/jinzhu/gorm v1.9.8 // indirect
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/miekg/pkcs11 v0.0.0-20170220202408-7283ca79f35e // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.34.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/spf13/viper v1.4.0 // indirect
	github.com/stretchr/objx v0.1.1 // indirect
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20150924051756-4e86f4367175 // indirect
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.42.0 // indirect
	gopkg.in/square/go-jose.v2 v2.3.0 // indirect
)
