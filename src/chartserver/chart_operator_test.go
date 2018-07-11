package chartserver

import (
	"testing"
)

func TestGetChartDetails(t *testing.T) {
	chartOpr := ChartOperator{}
	chartDetails, err := chartOpr.GetChartDetails(helmChartContent)
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
	infos, err := chartOpr.GetChartList(chartListContent)
	if err != nil {
		t.Fatal(err)
	}

	if len(infos) != 2 {
		t.Fatalf("Length of chart list should be 2, but we got %d now", len(infos))
	}

	foundHarbor := false
	for _, chart := range infos {
		if chart.Name == "harbor" {
			foundHarbor = true
			break
		}
	}

	if !foundHarbor {
		t.Fatal("Expect chart named with 'harbor' but got nothing")
	}
}
