package chart

import (
	"bytes"
	"errors"
	"fmt"
	helm_chart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"strings"
)

var (
	// Optr is a global chart operator instance
	Optr = NewOperator()
)

const (
	readmeFileName = "README.MD"
	valuesFileName = "VALUES.YAML"
)

// Operator ...
type Operator interface {
	// GetChartDetails parse the details from the provided content bytes
	GetDetails(content []byte) (*VersionDetails, error)
	// FetchLayer the content of layer under the repository
	GetData(content []byte) (*helm_chart.Chart, error)
}

var _ Operator = &operator{}

// ChartOperator is designed to process the contents of
// the specified chart version to get more details
type operator struct{}

// NewOperator returns an instance of the default chart opertaor
func NewOperator() Operator {
	return &operator{}
}

// GetDetails parse the details from the provided content bytes
func (cho *operator) GetDetails(content []byte) (*VersionDetails, error) {
	chartData, err := cho.GetData(content)
	if err != nil {
		return nil, err
	}

	// Parse the dependencies of chart
	depts := make([]*helm_chart.Dependency, 0)

	// for APIVersionV2, the dependency is in the Chart.yaml
	if chartData.Metadata.APIVersion == helm_chart.APIVersionV1 {
		depts = chartData.Metadata.Dependencies
	}

	var values map[string]interface{}
	files := make(map[string]string)
	// Parse values
	if chartData.Values != nil {
		readValue(values, "", chartData.Values)
	}

	// Append other files like 'README.md' 'values.yaml'
	for _, v := range chartData.Raw {
		if strings.ToUpper(v.Name) == readmeFileName {
			files[readmeFileName] = string(v.Data)
			break
		}
		if strings.ToUpper(v.Name) == valuesFileName {
			files[valuesFileName] = string(v.Data)
			break
		}
	}

	theChart := &VersionDetails{
		Dependencies: depts,
		Values:       values,
		Files:        files,
	}

	return theChart, nil
}

// GetData returns raw data of chart
func (cho *operator) GetData(content []byte) (*helm_chart.Chart, error) {
	if content == nil || len(content) == 0 {
		return nil, errors.New("zero content")
	}

	reader := bytes.NewReader(content)
	chartData, err := loader.LoadArchive(reader)
	if err != nil {
		return nil, err
	}

	return chartData, nil
}

// Recursively read value
func readValue(values map[string]interface{}, keyPrefix string, valueMap map[string]interface{}) {
	for key, value := range values {
		longKey := key
		if keyPrefix != "" {
			longKey = fmt.Sprintf("%s.%s", keyPrefix, key)
		}

		if subValues, ok := value.(map[string]interface{}); ok {
			readValue(subValues, longKey, valueMap)
		} else {
			valueMap[longKey] = value
		}
	}
}
