package chartserver

import (
	"testing"

	htesting "github.com/goharbor/harbor/src/testing"
)

func TestGetChartDetails(t *testing.T) {
	chartOpr := ChartOperator{}
	chartDetails, err := chartOpr.GetChartDetails(htesting.HelmChartContent)
	if err != nil {
		t.Fatal(err)
	}

	if len(chartDetails.Dependencies) == 0 {
		t.Fatal("At least 1 dependency exitsing, but we got 0 now")
	}

	if len(chartDetails.Values) == 0 {
		t.Fatal("At least 1 value existing, but we got 0 now")
	}

	if chartDetails.Values["adminserver.adminPassword"] != "Harbor12345" {
		t.Fatalf("The value of 'adminserver.adminPassword' should be 'Harbor12345' but we got '%s' now", chartDetails.Values["adminserver.adminPassword"])
	}
}

func TestGetChartList(t *testing.T) {
	chartOpr := ChartOperator{}
	infos, err := chartOpr.GetChartList(htesting.ChartListContent)
	if err != nil {
		t.Fatal(err)
	}

	if len(infos) != 2 {
		t.Fatalf("Length of chart list should be 2, but we got %d now", len(infos))
	}

	firstInSortedList := infos[0]
	if firstInSortedList.Name != "harbor" {
		t.Fatalf("Expect the fist item of the sorted list to be 'harbor' but got '%s'", firstInSortedList.Name)
	}

	if firstInSortedList.LatestVersion != "0.2.0" {
		t.Fatalf("Expect latest version '0.2.0' but got '%s'", firstInSortedList.LatestVersion)
	}
}
