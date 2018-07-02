package chartserver

import "github.com/kubernetes/helm/pkg/proto/hapi/chart"

//ChartOperator is designed to process the contents of
//the specified chart version to get more details
type ChartOperator struct{}

//GetChartDetails parse the details from the provided content bytes
func (cho *ChartOperator) GetChartDetails(content []byte) *chart.Chart {
	return nil
}
