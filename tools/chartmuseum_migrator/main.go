package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/goharbor/go-client/pkg/harbor"
	assistClient "github.com/goharbor/go-client/pkg/sdk/assist/client"
	"github.com/goharbor/go-client/pkg/sdk/assist/client/chart_repository"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
)

type HelmChart struct {
	Name    string
	Project string
	Version string
}

func (hc HelmChart) ChartFileName() string {
	return fmt.Sprintf("%s-%s.tgz", hc.Name, hc.Version)
}

const (
	fileMode       = 0o600
	helmBinaryPath = "helm"
	timeout        = 5 * time.Second
)

var (
	harborClientV2       *client.HarborAPI       //nolint:gochecknoglobals
	harborClientV2Assist *assistClient.HarborAPI //nolint:gochecknoglobals

	harborURL      string //nolint:gochecknoglobals
	harborUsername string //nolint:gochecknoglobals
	harborPassword string //nolint:gochecknoglobals
	harborHost     string //nolint:gochecknoglobals
)

func init() { //nolint:gochecknoinits
	initFlags()
	initHarborClients()
	initHarborHost()
}

func initFlags() {
	flag.StringVar(&harborURL, "url", "", "Harbor registry url")
	flag.StringVar(&harborUsername, "username", "", "Harbor registry username")
	flag.StringVar(&harborPassword, "password", "", "Harbor registry password")
	flag.Parse()

	if harborURL == "" {
		log.Fatal(errors.New("Missing required --url flag"))
	}

	if harborUsername == "" {
		log.Fatal(errors.New("Missing required --username flag"))
	}

	if harborPassword == "" {
		log.Fatal(errors.New("Missing required --password flag"))
	}
}

func initHarborClients() {
	config := &harbor.ClientSetConfig{
		URL:      harborURL,
		Insecure: false,
		Username: harborUsername,
		Password: harborPassword,
	}

	harborClientSet, err := harbor.NewClientSet(config)
	if err != nil {
		log.Fatal(err, errors.Wrap(err, "fail to create harbor client"))
	}

	harborClientV2 = harborClientSet.V2() // v2 client
	harborClientV2Assist = harborClientSet.Assist()

	// Check Harbor url and credentials are ok
	params := &project.ListProjectsParams{} //nolint:exhaustruct
	if _, err = harborClientV2.Project.ListProjects(context.Background(), params); err != nil {
		log.Fatal(errors.Wrap(err, "fail to contact Harbor registry, check your credentials"))
	}
}

func initHarborHost() {
	u, err := url.Parse(harborURL)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "fail to parse Harbor URL"))
	}

	harborHost = u.Host
}

func main() {
	if err := helmLogin(); err != nil {
		log.Fatal(errors.Wrapf(err, "fail to login to Helm"))
	}

	helmChartsToMigrate, err := getHarborChartmuseumCharts()
	if err != nil {
		log.Fatal(errors.Wrapf(err, "fail to retrieve helm charts to migrate"))
	}

	log.Printf("%d Helm charts to migrate from Chartmuseum to OCI", len(helmChartsToMigrate))

	bar := progressbar.Default(int64(len(helmChartsToMigrate)))
	errorCount := 0

	for _, helmChart := range helmChartsToMigrate {
		_ = bar.Add(1)

		if err := migrateChartFromChartmuseumToOCI(helmChart); err != nil {
			errorCount++

			log.Println(errors.Wrapf(err, "fail to migrate helm chart"))
		}
	}

	log.Printf("%d Helm charts successfully migrated from Chartmuseum to OCI", len(helmChartsToMigrate)-errorCount)
}

func helmLogin() error {
	cmd := exec.Command(helmBinaryPath, "registry", "login", "--username", harborUsername, "--password", harborPassword, harborURL) //nolint:lll

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "fail to execute helm push' command: %s", stdErr.String())
	}

	return nil
}

func getHarborChartmuseumCharts() ([]HelmChart, error) {
	helmCharts := make([]HelmChart, 0)

	params := &project.ListProjectsParams{} //nolint:exhaustruct

	projects, err := harborClientV2.Project.ListProjects(context.Background(), params)
	if err != nil {
		log.Fatal(errors.Wrap(err, "fail to list harbor projects"))
	}

	for _, harborProject := range projects.Payload {
		projectHelmCharts, err := getHarborProjectChartmuseumCharts(harborProject.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to migrate charts from project %s", harborProject.Name)
		}

		helmCharts = append(helmCharts, projectHelmCharts...)
	}

	return helmCharts, nil
}

func getHarborProjectChartmuseumCharts(projectName string) ([]HelmChart, error) {
	helmCharts := make([]HelmChart, 0)

	params := &chart_repository.GetChartrepoRepoChartsParams{Repo: projectName} //nolint:exhaustruct

	charts, err := harborClientV2Assist.ChartRepository.GetChartrepoRepoCharts(context.Background(), params)
	if err != nil {
		return nil, errors.Wrap(err, "fail to list harbor projects")
	}

	for _, chart := range charts.Payload {
		params := &chart_repository.GetChartrepoRepoChartsNameParams{ //nolint:exhaustruct
			Repo: projectName,
			Name: *chart.Name,
		}

		chartVersions, err := harborClientV2Assist.ChartRepository.GetChartrepoRepoChartsName(context.Background(), params)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to get chart %s", *chart.Name)
		}

		for _, chartVersion := range chartVersions.Payload {
			helmChart := HelmChart{
				Project: projectName,
				Name:    *chart.Name,
				Version: *chartVersion.Version,
			}

			helmCharts = append(helmCharts, helmChart)
		}
	}

	return helmCharts, nil
}

func migrateChartFromChartmuseumToOCI(helmChart HelmChart) error {
	if err := pullChartFromChartmuseum(helmChart); err != nil {
		return errors.Wrapf(err, "fail to pull chart from chartmuseum")
	}

	if err := pushChartToOCI(helmChart); err != nil {
		return errors.Wrapf(err, "fail to push chart to OCI")
	}

	if err := removeChartFile(helmChart); err != nil {
		return errors.Wrapf(err, "fail to remove chart file")
	}

	return nil
}

func pullChartFromChartmuseum(helmChart HelmChart) error {
	chartFileName := helmChart.ChartFileName()
	url := fmt.Sprintf("%s/chartrepo/%s/charts/%s", harborURL, helmChart.Project, chartFileName)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return errors.Wrapf(err, "fail to pull chart from Chartmuseum")
	}

	req.SetBasicAuth(harborUsername, harborPassword)

	httpClient := http.Client{Timeout: timeout} //nolint:exhaustruct

	res, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "fail to retrieve chart from chartmuseum")
	}

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("received status %d", res.StatusCode) //nolint:goerr113

		return errors.Wrap(err, "fail to retrieve chart from chartmuseum")
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Wrapf(err, "fail to read chart body")
	}

	err = os.WriteFile(chartFileName, resBody, fileMode)

	return errors.Wrapf(err, "fail to write chart file to disk")
}

func pushChartToOCI(helmChart HelmChart) error {
	repoURL := fmt.Sprintf("oci://%s/%s", harborHost, helmChart.Project)
	cmd := exec.Command(helmBinaryPath, "push", helmChart.ChartFileName(), repoURL) //nolint:gosec

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "fail to execute helm push' command: %s", stdErr.String())
	}

	return nil
}

func removeChartFile(helmChart HelmChart) error {
	chartFileName := helmChart.ChartFileName()

	err := os.Remove(chartFileName)

	return errors.Wrapf(err, "fail to delete file %s", chartFileName)
}
