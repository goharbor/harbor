package chart

import (
	"bytes"
	"errors"
	"fmt"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
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
	GetData(content []byte) (*chart.Chart, error)
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

	// Parse the requirements of chart
	requirements, err := chartutil.LoadRequirements(chartData)
	if err != nil {
		// If no requirements.yaml, return empty dependency list
		if _, ok := err.(chartutil.ErrNoRequirementsFile); ok {
			requirements = &chartutil.Requirements{
				Dependencies: make([]*chartutil.Dependency, 0),
			}
		} else {
			return nil, err
		}
	}

	var values map[string]interface{}
	files := make(map[string]string)
	// Parse values
	if chartData.Values != nil {
		values = parseRawValues([]byte(chartData.Values.GetRaw()))
		if len(values) > 0 {
			// Append values.yaml file
			files[valuesFileName] = chartData.Values.Raw
		}
	}

	// Append other files like 'README.md'
	for _, v := range chartData.GetFiles() {
		if v.TypeUrl == readmeFileName {
			files[readmeFileName] = string(v.GetValue())
			break
		}
	}

	theChart := &VersionDetails{
		Dependencies: requirements.Dependencies,
		Values:       values,
		Files:        files,
	}

	return theChart, nil
}

// GetData returns raw data of chart
func (cho *operator) GetData(content []byte) (*chart.Chart, error) {
	if content == nil || len(content) == 0 {
		return nil, errors.New("zero content")
	}

	reader := bytes.NewReader(content)
	chartData, err := chartutil.LoadArchive(reader)
	if err != nil {
		return nil, err
	}

	return chartData, nil
}

// Parse the raw values to value map
func parseRawValues(rawValue []byte) map[string]interface{} {
	valueMap := make(map[string]interface{})

	if len(rawValue) == 0 {
		return valueMap
	}

	values, err := chartutil.ReadValues(rawValue)
	if err != nil || len(values) == 0 {
		return valueMap
	}

	readValue(values, "", valueMap)

	return valueMap
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
