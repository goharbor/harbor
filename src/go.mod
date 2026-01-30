module github.com/goharbor/harbor/src

go 1.24.6

require (
	github.com/CloudNativeAI/model-spec v0.0.5
	github.com/FZambia/sentinel v1.1.0
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/aliyun/alibaba-cloud-sdk-go v1.63.107
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/aws/aws-sdk-go v1.55.8
	github.com/beego/beego/v2 v2.3.8
	github.com/beego/i18n v0.0.0-20140604031826-e87155e8f0c0
	github.com/bmatcuk/doublestar v1.3.4
	github.com/casbin/casbin v1.9.1
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/cloudevents/sdk-go/v2 v2.16.1
	github.com/coreos/go-oidc/v3 v3.15.0
	github.com/dghubble/sling v1.4.2
	github.com/docker/distribution v2.8.3+incompatible
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/go-asn1-ber/asn1-ber v1.5.8-0.20250403174932-29230038a667
	github.com/go-ldap/ldap/v3 v3.4.11
	github.com/go-openapi/errors v0.22.4
	github.com/go-openapi/loads v0.23.2
	github.com/go-openapi/runtime v0.29.2
	github.com/go-openapi/spec v0.22.1
	github.com/go-openapi/strfmt v0.25.0
	github.com/go-openapi/swag v0.23.1
	github.com/go-openapi/validate v0.25.1
	github.com/gocarina/gocsv v0.0.0-20210516172204-ca9e8a8ddea8
	github.com/gocraft/work v0.5.1
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/golang-migrate/migrate/v4 v4.19.0
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-containerregistry v0.20.2
	github.com/google/uuid v1.6.0
	github.com/gorilla/csrf v1.7.2
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/graph-gophers/dataloader v5.0.0+incompatible
	github.com/jackc/pgconn v1.14.3
	github.com/jackc/pgx/v4 v4.18.3
	github.com/jpillora/backoff v1.0.0
	github.com/ncw/swift v1.0.49 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/olekukonko/tablewriter v0.0.5
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.22.0
	github.com/prometheus/client_model v0.6.2
	github.com/redis/go-redis/v9 v9.17.2
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.11.1
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible
	github.com/vmihailenco/msgpack/v5 v5.4.1
	github.com/volcengine/volcengine-go-sdk v1.1.44
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.63.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.34.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
	go.pinniped.dev v0.37.0
	go.uber.org/ratelimit v0.3.1
	golang.org/x/crypto v0.45.0
	golang.org/x/net v0.47.0
	golang.org/x/oauth2 v0.28.0
	golang.org/x/sync v0.18.0
	golang.org/x/text v0.31.0
	golang.org/x/time v0.13.0
	gopkg.in/h2non/gock.v1 v1.1.2
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.18.5
	k8s.io/api v0.34.1
	k8s.io/apimachinery v0.34.2
	k8s.io/client-go v0.34.1
	sigs.k8s.io/yaml v1.6.0
)

require (
	cloud.google.com/go/compute/metadata v0.5.2 // indirect
	github.com/Azure/azure-sdk-for-go v37.2.0+incompatible // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.27 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.20 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/Unknwon/goconfig v0.0.0-20160216183935-5f601ca6ef4d // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.14.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/denverdino/aliyungo v0.0.0-20191227032621-df38c6fa730c // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dnaeon/go-vcr v1.2.0 // indirect
	github.com/docker/cli v27.1.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.2 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.24.1 // indirect
	github.com/go-openapi/jsonpointer v0.22.1 // indirect
	github.com/go-openapi/jsonreference v0.21.3 // indirect
	github.com/go-openapi/swag/conv v0.25.1 // indirect
	github.com/go-openapi/swag/fileutils v0.25.1 // indirect
	github.com/go-openapi/swag/jsonname v0.25.1 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.1 // indirect
	github.com/go-openapi/swag/loading v0.25.1 // indirect
	github.com/go-openapi/swag/mangling v0.25.1 // indirect
	github.com/go-openapi/swag/stringutils v0.25.1 // indirect
	github.com/go-openapi/swag/typeutils v0.25.1 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.25.1 // indirect
	github.com/h2non/parth v0.0.0-20190131123155-b4df798d6542 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/robfig/cron v1.0.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smartystreets/goconvey v1.8.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/vbatts/tar-split v0.11.3 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/volcengine/volc-sdk-golang v1.0.23 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.mongodb.org/mongo-driver v1.17.6 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0 // indirect
	google.golang.org/api v0.171.0 // indirect
	google.golang.org/cloud v0.0.0-20151119220103-975617b05ea8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/grpc v1.69.4 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/docker/distribution => github.com/distribution/distribution v2.8.2+incompatible
	github.com/gocraft/work => github.com/goharbor/work v0.5.1-patch
	github.com/goharbor/harbor => ../
	github.com/gomodule/redigo => github.com/gomodule/redigo v1.9.2
	google.golang.org/api => google.golang.org/api v0.0.0-20160322025152-9bf6e6e569ff
)
