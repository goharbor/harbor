package helmhub

import "fmt"

const (
	baseURL    = "https://hub.helm.sh"
	listCharts = "/api/chartsvc/v1/charts"
)

func listVersions(chartName string) string {
	return fmt.Sprintf("/api/chartsvc/v1/charts/%s/versions", chartName)
}
