package endpoints

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	separator      = "."
	openPrefix     = "open"
	endpointSuffix = separator + "volcengineapi.com"
)

var defaultEndpoint = openPrefix + endpointSuffix

type RegionEndpointMap map[string]string

type ServiceEndpointInfo struct {
	Service           string `yaml:"Service"`
	IsGlobal          bool   `yaml:"IsGlobal"`
	GlobalEndpoint    string `yaml:"GlobalEndpoint"`
	RegionEndpointMap `yaml:"RegionEndpointMap"`
}

type FileEndpointConfigResolver struct {
	Path           string
	EndpointConfig map[string]*ServiceEndpointInfo
}

func (endpointResolver *FileEndpointConfigResolver) Load() error {
	file, err := ioutil.ReadFile(endpointResolver.Path)
	if err != nil {
		return err
	}
	config := make(map[string]*ServiceEndpointInfo)

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return err
	}
	endpointResolver.EndpointConfig = config
	return nil
}

func (endpointResolver *FileEndpointConfigResolver) EndpointFor(service, region string, opts ...func(*Options)) (ResolvedEndpoint, error) {
	result := ResolvedEndpoint{}
	result.URL = defaultEndpoint
	defaultEndpointInfo, sExist := endpointResolver.EndpointConfig[service]
	if !sExist {
		return result, nil
	}

	isGlobal := defaultEndpointInfo.IsGlobal
	if isGlobal {
		result.URL = defaultEndpointInfo.GlobalEndpoint
		return result, nil
	}

	regionEndpointMp := defaultEndpointInfo.RegionEndpointMap
	regionEndpointStr, rExist := regionEndpointMp[region]
	if !rExist {
		return result, nil
	}

	result.URL = regionEndpointStr
	return result, nil
}
