package volcengine

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"net/http"
	"time"

	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/custom"
	"github.com/volcengine/volcengine-go-sdk/volcengine/endpoints"
)

// UseServiceDefaultRetries instructs the config to use the service's own
// default number of retries. This will be the default action if
// Config.MaxRetries is nil also.
const UseServiceDefaultRetries = -1

// RequestRetryer is an alias for a type that implements the request.Retryer
// interface.
type RequestRetryer interface{}

// A Config provides service configuration for service clients. By default,
// all clients will use the defaults.DefaultConfig structure.
type Config struct {
	// Enables verbose error printing of all credential chain errors.
	// Should be used when wanting to see all errors while attempting to
	// retrieve credentials.
	CredentialsChainVerboseErrors *bool

	// The credentials object to use when signing requests. Defaults to a
	// chain of credential providers to search for credentials in environment
	// variables, shared credential file, and EC2 Instance Roles.
	Credentials *credentials.Credentials

	// An optional endpoint URL (hostname only or fully qualified URI)
	// that overrides the default generated endpoint for a client. Set this
	// to `""` to use the default generated endpoint.
	//
	// Note: You must still provide a `Region` value when specifying an
	// endpoint for a client.
	Endpoint *string

	// The resolver to use for looking up endpoints for volcengine service clients
	// to use based on region.
	EndpointResolver endpoints.Resolver

	// EnforceShouldRetryCheck is used in the AfterRetryHandler to always call
	// ShouldRetry regardless of whether or not if request.Retryable is set.
	// This will utilize ShouldRetry method of custom retryers. If EnforceShouldRetryCheck
	// is not set, then ShouldRetry will only be called if request.Retryable is nil.
	// Proper handling of the request.Retryable field is important when setting this field.
	EnforceShouldRetryCheck *bool

	// The region to send requests to. This parameter is required and must
	// be configured globally or on a per-client basis unless otherwise
	// noted. A full list of regions is found in the "Regions and Endpoints"
	// document.
	//
	// Regions and Endpoints.
	Region *string

	// Set this to `true` to disable SSL when sending requests. Defaults
	// to `false`.
	DisableSSL *bool

	// The HTTP client to use when sending requests. Defaults to
	// `http.DefaultClient`.
	HTTPClient *http.Client

	// An integer value representing the logging level. The default log level
	// is zero (LogOff), which represents no logging. To enable logging set
	// to a LogLevel Value.
	LogLevel *LogLevelType

	// The logger writer interface to write logging messages to. Defaults to
	// standard out.
	Logger Logger

	// The maximum number of times that a request will be retried for failures.
	// Defaults to -1, which defers the max retry setting to the service
	// specific configuration.
	MaxRetries *int

	// Retryer guides how HTTP requests should be retried in case of
	// recoverable failures.
	//
	// When nil or the value does not implement the request.Retryer interface,
	// the client.DefaultRetryer will be used.
	//
	// When both Retryer and MaxRetries are non-nil, the former is used and
	// the latter ignored.
	//
	// To set the Retryer field in a type-safe manner and with chaining, use
	// the request.WithRetryer helper function:
	//
	//   cfg := request.WithRetryer(volcengine.NewConfig(), myRetryer)
	//
	Retryer RequestRetryer

	// Disables semantic parameter validation, which validates input for
	// missing required fields and/or other semantic request input errors
	// Temporary notes by xuyaming@bytedance.com because some validate field is relation.
	//DisableParamValidation *bool

	// Disables the computation of request and response checksums, e.g.,
	//DisableComputeChecksums *bool

	// Set this to `true` to enable S3 Accelerate feature. For all operations
	// compatible with S3 Accelerate will use the accelerate endpoint for
	// requests. Requests not compatible will fall back to normal S3 requests.
	//
	// The bucket must be enable for accelerate to be used with S3 client with
	// accelerate enabled. If the bucket is not enabled for accelerate an error
	// will be returned. The bucket name must be DNS compatible to also work
	// with accelerate.
	//S3UseAccelerate *bool

	// S3DisableContentMD5Validation config option is temporarily disabled,
	// For S3 GetObject API calls, #1837.
	//
	// Set this to `true` to disable the S3 service client from automatically
	// adding the ContentMD5 to S3 Object Put and Upload API calls. This option
	// will also disable the SDK from performing object ContentMD5 validation
	// on GetObject API calls.
	//S3DisableContentMD5Validation *bool

	// Set this to `true` to disable the EC2Metadata client from overriding the
	// default http.Client's Timeout. This is helpful if you do not want the
	// EC2Metadata client to create a new http.Client. This options is only
	// meaningful if you're not already using a custom HTTP client with the
	// SDK. Enabled by default.
	//
	// Must be set and provided to the session.NewSession() in order to disable
	// the EC2Metadata overriding the timeout for default credentials chain.
	//
	// Example:
	//    sess := session.Must(session.NewSession(volcengine.NewConfig()
	//       .WithEC2MetadataDiableTimeoutOverride(true)))
	//
	//    svc := s3.New(sess)
	//
	//EC2MetadataDisableTimeoutOverride *bool

	// Instructs the endpoint to be generated for a service client to
	// be the dual stack endpoint. The dual stack endpoint will support
	// both IPv4 and IPv6 addressing.
	//
	// Setting this for a service which does not support dual stack will fail
	// to make requets. It is not recommended to set this value on the session
	// as it will apply to all service clients created with the session. Even
	// services which don't support dual stack endpoints.
	//
	// If the Endpoint config value is also provided the UseDualStack flag
	// will be ignored.
	//
	// Only supported with.
	//
	//     sess := session.Must(session.NewSession())
	//
	//     svc := s3.New(sess, &volcengine.Config{
	//         UseDualStack: volcengine.Bool(true),
	//     })
	//UseDualStack *bool

	// SleepDelay is an override for the func the SDK will call when sleeping
	// during the lifecycle of a request. Specifically this will be used for
	// request delays. This value should only be used for testing. To adjust
	// the delay of a request see the volcengine/client.DefaultRetryer and
	// volcengine/request.Retryer.
	//
	// SleepDelay will prevent any Context from being used for canceling retry
	// delay of an API operation. It is recommended to not use SleepDelay at all
	// and specify a Retryer instead.
	SleepDelay func(time.Duration)

	// DisableRestProtocolURICleaning will not clean the URL path when making rest protocol requests.
	// Will default to false. This would only be used for empty directory names in s3 requests.
	//
	// Example:
	//    sess := session.Must(session.NewSession(&volcengine.Config{
	//         DisableRestProtocolURICleaning: volcengine.Bool(true),
	//    }))
	//
	//    svc := s3.New(sess)
	//    out, err := svc.GetObject(&s3.GetObjectInput {
	//    	Bucket: volcengine.String("bucketname"),
	//    	Key: volcengine.String("//foo//bar//moo"),
	//    })
	DisableRestProtocolURICleaning *bool

	// EnableEndpointDiscovery will allow for endpoint discovery on operations that
	// have the definition in its model. By default, endpoint discovery is off.
	//
	// Example:
	//    sess := session.Must(session.NewSession(&volcengine.Config{
	//         EnableEndpointDiscovery: volcengine.Bool(true),
	//    }))
	//
	//    svc := s3.New(sess)
	//    out, err := svc.GetObject(&s3.GetObjectInput {
	//    	Bucket: volcengine.String("bucketname"),
	//    	Key: volcengine.String("/foo/bar/moo"),
	//    })
	//EnableEndpointDiscovery *bool

	// DisableEndpointHostPrefix will disable the SDK's behavior of prefixing
	// request endpoint hosts with modeled information.
	//
	// Disabling this feature is useful when you want to use local endpoints
	// for testing that do not support the modeled host prefix pattern.
	//DisableEndpointHostPrefix *bool

	LogSensitives []string

	DynamicCredentials custom.DynamicCredentials

	DynamicCredentialsIncludeError custom.DynamicCredentialsIncludeError

	LogAccount custom.LogAccount

	ExtendHttpRequest custom.ExtendHttpRequest

	ExtendHttpRequestWithMeta custom.ExtendHttpRequestWithMeta

	ExtraHttpParameters custom.ExtraHttpParameters

	ExtraHttpParametersWithMeta custom.ExtraHttpParametersWithMeta

	ExtraHttpJsonBody custom.ExtraHttpJsonBody

	CustomerUnmarshalError custom.CustomerUnmarshalError

	CustomerUnmarshalData custom.CustomerUnmarshalData

	ExtendContextWithMeta custom.ExtendContextWithMeta

	ExtraUserAgent *string

	Interceptors []custom.SdkInterceptor

	SimpleError *bool

	ForceJsonNumberDecode custom.ForceJsonNumberDecode

	EndpointConfigState *bool

	EndpointConfigPath *string
}

// NewConfig returns a new Config pointer that can be chained with builder
// methods to set multiple configuration values inline without using pointers.
//
//     // Create Session with MaxRetries configuration to be shared by multiple
//     // service clients.
//     sess := session.Must(session.NewSession(volcengine.NewConfig().
//         WithMaxRetries(3),
//     ))
//
//     // Create S3 service client with a specific Region.
//     svc := s3.New(sess, volcengine.NewConfig().
//         WithRegion("us-west-2"),
//     )
func NewConfig() *Config {
	return &Config{}
}

func (c *Config) AddInterceptor(interceptor custom.SdkInterceptor) *Config {
	c.Interceptors = append(c.Interceptors, interceptor)
	return c
}

// WithCredentialsChainVerboseErrors sets a config verbose errors boolean and returning
// a Config pointer.
func (c *Config) WithCredentialsChainVerboseErrors(verboseErrs bool) *Config {
	c.CredentialsChainVerboseErrors = &verboseErrs
	return c
}

// WithCredentials sets a config Credentials value returning a Config pointer
// for chaining.
func (c *Config) WithCredentials(creds *credentials.Credentials) *Config {
	c.Credentials = creds
	return c
}

// WithAkSk sets a config Credentials value returning a Config pointer
// for chaining.
func (c *Config) WithAkSk(ak, sk string) *Config {
	c.Credentials = credentials.NewStaticCredentials(ak, sk, "")
	return c
}

// WithEndpoint sets a config Endpoint value returning a Config pointer for
// chaining.
func (c *Config) WithEndpoint(endpoint string) *Config {
	c.Endpoint = &endpoint
	return c
}

func (c *Config) WithSimpleError(simpleError bool) *Config {
	c.SimpleError = &simpleError
	return c
}

func (c *Config) WithLogSensitives(sensitives []string) *Config {
	c.LogSensitives = sensitives
	return c
}

func (c *Config) WithExtendHttpRequest(extendHttpRequest custom.ExtendHttpRequest) *Config {
	c.ExtendHttpRequest = extendHttpRequest
	return c
}

func (c *Config) WithExtendHttpRequestWithMeta(extendHttpRequestWithMeta custom.ExtendHttpRequestWithMeta) *Config {
	c.ExtendHttpRequestWithMeta = extendHttpRequestWithMeta
	return c
}

func (c *Config) WithExtendContextWithMeta(extendContextWithMeta custom.ExtendContextWithMeta) *Config {
	c.ExtendContextWithMeta = extendContextWithMeta
	return c
}

func (c *Config) WithExtraHttpParameters(extraHttpParameters custom.ExtraHttpParameters) *Config {
	c.ExtraHttpParameters = extraHttpParameters
	return c
}

func (c *Config) WithExtraHttpParametersWithMeta(extraHttpParametersWithMeta custom.ExtraHttpParametersWithMeta) *Config {
	c.ExtraHttpParametersWithMeta = extraHttpParametersWithMeta
	return c
}

func (c *Config) WithExtraHttpJsonBody(extraHttpJsonBody custom.ExtraHttpJsonBody) *Config {
	c.ExtraHttpJsonBody = extraHttpJsonBody
	return c
}

func (c *Config) WithExtraUserAgent(extra *string) *Config {
	c.ExtraUserAgent = extra
	return c
}

func (c *Config) WithLogAccount(account custom.LogAccount) *Config {
	c.LogAccount = account
	return c
}

func (c *Config) WithDynamicCredentials(f custom.DynamicCredentials) *Config {
	c.DynamicCredentials = f
	return c
}

// WithDynamicCredentialsIncludeError sets a config DynamicCredentialsIncludeError value returning a Config pointer for
// chaining.
func (c *Config) WithDynamicCredentialsIncludeError(f custom.DynamicCredentialsIncludeError) *Config {
	c.DynamicCredentialsIncludeError = f
	return c
}

func (c *Config) WithCustomerUnmarshalError(f custom.CustomerUnmarshalError) *Config {
	c.CustomerUnmarshalError = f
	return c
}

func (c *Config) WithCustomerUnmarshalData(f custom.CustomerUnmarshalData) *Config {
	c.CustomerUnmarshalData = f
	return c
}

func (c *Config) WithForceJsonNumberDecode(f custom.ForceJsonNumberDecode) *Config {
	c.ForceJsonNumberDecode = f
	return c
}

// WithEndpointResolver sets a config EndpointResolver value returning a
// Config pointer for chaining.
func (c *Config) WithEndpointResolver(resolver endpoints.Resolver) *Config {
	c.EndpointResolver = resolver
	return c
}

// WithRegion sets a config Region value returning a Config pointer for
// chaining.
func (c *Config) WithRegion(region string) *Config {
	c.Region = &region
	return c
}

// WithDisableSSL sets a config DisableSSL value returning a Config pointer
// for chaining.
func (c *Config) WithDisableSSL(disable bool) *Config {
	c.DisableSSL = &disable
	return c
}

// WithHTTPClient sets a config HTTPClient value returning a Config pointer
// for chaining.
func (c *Config) WithHTTPClient(client *http.Client) *Config {
	c.HTTPClient = client
	return c
}

// WithMaxRetries sets a config MaxRetries value returning a Config pointer
// for chaining.
func (c *Config) WithMaxRetries(max int) *Config {
	c.MaxRetries = &max
	return c
}

// WithDisableParamValidation sets a config DisableParamValidation value
// returning a Config pointer for chaining
// Temporary notes by xuyaming@bytedance.com because some validate field is relation.
//func (c *Config) WithDisableParamValidation(disable bool) *Config {
//	c.DisableParamValidation = &disable
//	return c
//}

// WithDisableComputeChecksums sets a config DisableComputeChecksums value
// returning a Config pointer for chaining.
//func (c *Config) WithDisableComputeChecksums(disable bool) *Config {
//	c.DisableComputeChecksums = &disable
//	return c
//}

// WithLogLevel sets a config LogLevel value returning a Config pointer for
// chaining.
func (c *Config) WithLogLevel(level LogLevelType) *Config {
	c.LogLevel = &level
	return c
}

// WithLogger sets a config Logger value returning a Config pointer for
// chaining.
func (c *Config) WithLogger(logger Logger) *Config {
	c.Logger = logger
	return c
}

// WithS3UseAccelerate sets a config S3UseAccelerate value returning a Config
// pointer for chaining.
//func (c *Config) WithS3UseAccelerate(enable bool) *Config {
//	c.S3UseAccelerate = &enable
//	return c
//
//}

// WithS3DisableContentMD5Validation sets a config
// S3DisableContentMD5Validation value returning a Config pointer for chaining.
//func (c *Config) WithS3DisableContentMD5Validation(enable bool) *Config {
//	c.S3DisableContentMD5Validation = &enable
//	return c
//
//}

// WithUseDualStack sets a config UseDualStack value returning a Config
// pointer for chaining.
//func (c *Config) WithUseDualStack(enable bool) *Config {
//	c.UseDualStack = &enable
//	return c
//}

// WithEC2MetadataDisableTimeoutOverride sets a config EC2MetadataDisableTimeoutOverride value
// returning a Config pointer for chaining.
//func (c *Config) WithEC2MetadataDisableTimeoutOverride(enable bool) *Config {
//	c.EC2MetadataDisableTimeoutOverride = &enable
//	return c
//}

// WithSleepDelay overrides the function used to sleep while waiting for the
// next retry. Defaults to time.Sleep.
func (c *Config) WithSleepDelay(fn func(time.Duration)) *Config {
	c.SleepDelay = fn
	return c
}

// WithEndpointDiscovery will set whether or not to use endpoint discovery.
//func (c *Config) WithEndpointDiscovery(t bool) *Config {
//	c.EnableEndpointDiscovery = &t
//	return c
//}

// WithDisableEndpointHostPrefix will set whether or not to use modeled host prefix
// when making requests.
//func (c *Config) WithDisableEndpointHostPrefix(t bool) *Config {
//	c.DisableEndpointHostPrefix = &t
//	return c
//}
// WithEndpointConfigState will set whether or not to use FileEndpointResolver
func (c *Config) WithEndpointConfigState(t bool) *Config {
	c.EndpointConfigState = &t
	return c
}

// WithEndpointConfigPath will set  fileEndpointResolver config path . This takes effect when EndpointConfigState is true.
func (c *Config) WithEndpointConfigPath(path string) *Config {
	c.EndpointConfigPath = &path
	return c
}

// MergeIn merges the passed in configs into the existing config object.
func (c *Config) MergeIn(cfgs ...*Config) {
	for _, other := range cfgs {
		mergeInConfig(c, other)
	}
}

func mergeInConfig(dst *Config, other *Config) {
	if other == nil {
		return
	}

	if other.CredentialsChainVerboseErrors != nil {
		dst.CredentialsChainVerboseErrors = other.CredentialsChainVerboseErrors
	}

	if other.Credentials != nil {
		dst.Credentials = other.Credentials
	}

	if other.Endpoint != nil {
		dst.Endpoint = other.Endpoint
	}

	if other.EndpointResolver != nil {
		dst.EndpointResolver = other.EndpointResolver
	}

	if other.Region != nil {
		dst.Region = other.Region
	}

	if other.DisableSSL != nil {
		dst.DisableSSL = other.DisableSSL
	}

	if other.HTTPClient != nil {
		dst.HTTPClient = other.HTTPClient
	}

	if other.LogLevel != nil {
		dst.LogLevel = other.LogLevel
	}

	if other.Logger != nil {
		dst.Logger = other.Logger
	}

	if other.MaxRetries != nil {
		dst.MaxRetries = other.MaxRetries
	}

	if other.Retryer != nil {
		dst.Retryer = other.Retryer
	}

	if other.ForceJsonNumberDecode != nil {
		dst.ForceJsonNumberDecode = other.ForceJsonNumberDecode
	}

	if other.ExtendContextWithMeta != nil {
		dst.ExtendContextWithMeta = other.ExtendContextWithMeta
	}
	// Temporary notes by xuyaming@bytedance.com because some validate field is relation.
	//if other.DisableParamValidation != nil {
	//	dst.DisableParamValidation = other.DisableParamValidation
	//}

	//if other.DisableComputeChecksums != nil {
	//	dst.DisableComputeChecksums = other.DisableComputeChecksums
	//}

	//if other.S3UseAccelerate != nil {
	//	dst.S3UseAccelerate = other.S3UseAccelerate
	//}
	//
	//if other.S3DisableContentMD5Validation != nil {
	//	dst.S3DisableContentMD5Validation = other.S3DisableContentMD5Validation
	//}
	//
	//if other.UseDualStack != nil {
	//	dst.UseDualStack = other.UseDualStack
	//}
	//
	//if other.EC2MetadataDisableTimeoutOverride != nil {
	//	dst.EC2MetadataDisableTimeoutOverride = other.EC2MetadataDisableTimeoutOverride
	//}
	//
	if other.SleepDelay != nil {
		dst.SleepDelay = other.SleepDelay
	}
	//
	if other.DisableRestProtocolURICleaning != nil {
		dst.DisableRestProtocolURICleaning = other.DisableRestProtocolURICleaning
	}
	//
	//if other.EnforceShouldRetryCheck != nil {
	//	dst.EnforceShouldRetryCheck = other.EnforceShouldRetryCheck
	//}
	//
	//if other.EnableEndpointDiscovery != nil {
	//	dst.EnableEndpointDiscovery = other.EnableEndpointDiscovery
	//}
	//
	//if other.DisableEndpointHostPrefix != nil {
	//	dst.DisableEndpointHostPrefix = other.DisableEndpointHostPrefix
	//}

	if other.LogSensitives != nil {
		dst.LogSensitives = other.LogSensitives
	}

	if other.LogAccount != nil {
		dst.LogAccount = other.LogAccount
	}

	if other.DynamicCredentials != nil {
		dst.DynamicCredentials = other.DynamicCredentials
	}

	if other.DynamicCredentialsIncludeError != nil {
		dst.DynamicCredentialsIncludeError = other.DynamicCredentialsIncludeError
	}

	if other.ExtendHttpRequest != nil {
		dst.ExtendHttpRequest = other.ExtendHttpRequest
	}

	if other.ExtendHttpRequestWithMeta != nil {
		dst.ExtendHttpRequestWithMeta = other.ExtendHttpRequestWithMeta
	}

	if other.ExtraHttpParameters != nil {
		dst.ExtraHttpParameters = other.ExtraHttpParameters
	}

	if other.ExtraHttpParametersWithMeta != nil {
		dst.ExtraHttpParametersWithMeta = other.ExtraHttpParametersWithMeta
	}

	if other.ExtraUserAgent != nil {
		dst.ExtraUserAgent = other.ExtraUserAgent
	}

	if other.SimpleError != nil {
		dst.SimpleError = other.SimpleError
	}

	if other.ExtraHttpJsonBody != nil {
		dst.ExtraHttpJsonBody = other.ExtraHttpJsonBody
	}

	if other.CustomerUnmarshalError != nil {
		dst.CustomerUnmarshalError = other.CustomerUnmarshalError
	}

	if other.CustomerUnmarshalData != nil {
		dst.CustomerUnmarshalData = other.CustomerUnmarshalData
	}

	if other.EndpointConfigState != nil {
		dst.EndpointConfigState = other.EndpointConfigState
	}

	if other.EndpointConfigPath != nil {
		dst.EndpointConfigPath = other.EndpointConfigPath
	}

	dst.Interceptors = other.Interceptors
}

// Copy will return a shallow copy of the Config object. If any additional
// configurations are provided they will be merged into the new config returned.
func (c *Config) Copy(cfgs ...*Config) *Config {
	dst := &Config{}
	dst.MergeIn(c)

	for _, cfg := range cfgs {
		dst.MergeIn(cfg)
	}

	return dst
}
