package volcengineutil

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

type Endpoint struct {
	//UseSSL bool
	//Locate bool
	//UseInternal                 bool
	CustomerEndpoint string
	//CustomerDomainIgnoreService bool

}

func NewEndpoint() *Endpoint {
	return &Endpoint{}
}

func (c *Endpoint) WithCustomerEndpoint(customerEndpoint string) *Endpoint {
	c.CustomerEndpoint = customerEndpoint
	return c
}

type ServiceInfo struct {
	Service string
	Region  string
}

func (c *Endpoint) GetEndpoint() string {
	if c.CustomerEndpoint != "" {
		return c.CustomerEndpoint
	} else {
		return endpoint
	}
}

const (
	separator      = "."
	openPrefix     = "open"
	endpointSuffix = separator + "volcengineapi.com"
)

var endpoint = openPrefix + endpointSuffix

type RegionEndpointMap map[string]string

type ServiceEndpointInfo struct {
	Service         string
	IsGlobal        bool
	GlobalEndpoint  string
	DefaultEndpoint string
	RegionEndpointMap
}

const (
	regionCodeCNBeijingAutoDriving = "cn-beijing-autodriving"
)

var defaultEndpoint = map[string]*ServiceEndpointInfo{
	"vke": {
		Service:         "vke",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "vke" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"Redis": {
		Service:         "Redis",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "redis" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"privatelink": {
		Service:         "privatelink",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "privatelink" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"vpc": {
		Service:         "vpc",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "vpc" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"RocketMQ": {
		Service:         "RocketMQ",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "rocketmq" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"Kafka": {
		Service:         "Kafka",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "kafka" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"rds_mysql": {
		Service:         "rds_mysql",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "rds-mysql" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"rds_postgresql": {
		Service:         "rds_postgresql",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "rds-postgresql" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"storage_ebs": {
		Service:         "storage_ebs",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "storage-ebs" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"clb": {
		Service:         "clb",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "clb" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"ecs": {
		Service:         "ecs",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "ecs" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"alb": {
		Service:         "alb",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "alb" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"FileNAS": {
		Service:         "FileNAS",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "filenas" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"configcenter": {
		Service:         "configcenter",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "configcenter" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"cr": {
		Service:         "cr",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "cr" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"sts": {
		Service:         "sts",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "sts" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"mongodb": {
		Service:         "mongodb",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "mongodb" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"transitrouter": {
		Service:         "transitrouter",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "transitrouter" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"Volc_Observe": {
		Service:         "Volc_Observe",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "volc-observe" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"dms": {
		Service:         "dms",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "dms" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"auto_scaling": {
		Service:         "auto_scaling",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "auto-scaling" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"directconnect": {
		Service:         "directconnect",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "directconnect" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
	"kms": {
		Service:         "kms",
		IsGlobal:        false,
		GlobalEndpoint:  "",
		DefaultEndpoint: endpoint,
		RegionEndpointMap: RegionEndpointMap{
			regionCodeCNBeijingAutoDriving: "kms" + separator + regionCodeCNBeijingAutoDriving + endpointSuffix,
		},
	},
}

// GetDefaultEndpointByServiceInfo retrieves the default endpoint for a given service and region.
//
// This function takes in the service name and region code, and returns a pointer to the default
// endpoint string. It checks if the service has a global endpoint or a region-specific endpoint.
// If neither is found, it returns a pointer to the default endpoint.
//
// Parameters:
// - service: The name of the service for which to retrieve the endpoint.
// - regionCode: The region code to look up the region-specific endpoint.
//
// Returns:
// - *string: A pointer to the endpoint string. It could be a global endpoint, a region-specific
// endpoint, or a default endpoint if the specified service or region does not have a defined endpoint.
//
// Example:
//   endpoint := GetDefaultEndpointByServiceInfo("exampleService", "cn-beijing")
//
// Note: Ensure the `defaultEndpoint` map is properly populated with service and region endpoint
// information before calling this function.
func GetDefaultEndpointByServiceInfo(service string, regionCode string) *string {
	resultEndpoint := endpoint
	defaultEndpointInfo, sExist := defaultEndpoint[service]
	if !sExist {
		return &resultEndpoint
	}

	isGlobal := defaultEndpointInfo.IsGlobal
	if isGlobal {
		if len(defaultEndpointInfo.GlobalEndpoint) > 0 {
			resultEndpoint = defaultEndpointInfo.GlobalEndpoint
			return &resultEndpoint
		}
	} else {
		regionEndpointMp := defaultEndpointInfo.RegionEndpointMap
		regionEndpointStr, rExist := regionEndpointMp[regionCode]
		if rExist {
			resultEndpoint = regionEndpointStr
			return &resultEndpoint
		}
	}

	if len(defaultEndpointInfo.DefaultEndpoint) > 0 {
		resultEndpoint = defaultEndpointInfo.DefaultEndpoint
		return &resultEndpoint
	}
	return &resultEndpoint
}
