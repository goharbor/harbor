package chartserver

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

//UtilityHandler provides utility methods
type UtilityHandler struct {
	//Parse and process the chart version to provide required info data
	chartOperator *ChartOperator

	//HTTP client used to call the realted APIs of the backend chart repositories
	apiClient *ChartClient

	//Point to the url of the backend server
	backendServerAddress *url.URL
}

//GetChartsByNs gets the chart list under the namespace
func (uh *UtilityHandler) GetChartsByNs(namespace string) ([]*ChartInfo, error) {
	if len(strings.TrimSpace(namespace)) == 0 {
		return nil, errors.New("empty namespace when getting chart list")
	}

	path := fmt.Sprintf("/api/%s/charts", namespace)
	url := fmt.Sprintf("%s%s", uh.backendServerAddress.String(), path)

	content, err := uh.apiClient.GetContent(url)
	if err != nil {
		return nil, err
	}

	return uh.chartOperator.GetChartList(content)
}
