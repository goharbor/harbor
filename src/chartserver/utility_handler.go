package chartserver

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	helm_repo "k8s.io/helm/pkg/repo"
)

const (
	maxDeletionThreads = 10
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

//DeleteChart deletes all the chart versions of the specified chart under the namespace.
func (uh *UtilityHandler) DeleteChart(namespace, chartName string) error {
	if len(strings.TrimSpace(namespace)) == 0 {
		return errors.New("empty namespace when deleting chart")
	}

	if len(strings.TrimSpace(chartName)) == 0 {
		return errors.New("empty chart name when deleting chart")
	}

	path := fmt.Sprintf("/api/%s/charts/%s", namespace, chartName)
	url := fmt.Sprintf("%s%s", uh.backendServerAddress.String(), path)

	content, err := uh.apiClient.GetContent(url)
	if err != nil {
		return err
	}

	allVersions, err := uh.chartOperator.GetChartVersions(content)
	if err != nil {
		return err
	}

	//Let's delete the versions in parallel
	//The number of goroutine is controlled by the const maxDeletionThreads
	qSize := len(allVersions)
	if qSize > maxDeletionThreads {
		qSize = maxDeletionThreads
	}
	tokenQueue := make(chan struct{}, qSize)
	errChan := make(chan error, 1)
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(allVersions))

	//Append initial tokens
	for i := 0; i < qSize; i++ {
		tokenQueue <- struct{}{}
	}

	//Collect errors
	errs := make([]error, 0)
	errWrapper := make(chan error, 1)
	go func() {
		defer func() {
			//pass to the out func
			if len(errs) > 0 {
				errWrapper <- fmt.Errorf("%v", errs)
			}
			close(errWrapper)
		}()

		for deletionErr := range errChan {
			errs = append(errs, deletionErr)
		}
	}()

	//Schedule deletion tasks
	for _, deletingVersion := range allVersions {
		//Apply for token first
		//If no available token, pending here
		<-tokenQueue

		//Got one token
		go func(deletingVersion *helm_repo.ChartVersion) {
			defer func() {
				//return the token back
				tokenQueue <- struct{}{}

				//done
				waitGroup.Done()
			}()

			if err := uh.deleteChartVersion(namespace, chartName, deletingVersion.GetVersion()); err != nil {
				errChan <- err
			}
		}(deletingVersion)
	}

	//Wait all goroutines are done
	waitGroup.Wait()
	//Safe to quit error collection goroutine
	close(errChan)

	err = <-errWrapper

	return err
}

//deleteChartVersion deletes the specified chart version
func (uh *UtilityHandler) deleteChartVersion(namespace, chartName, version string) error {
	path := fmt.Sprintf("/api/%s/charts/%s/%s", namespace, chartName, version)
	url := fmt.Sprintf("%s%s", uh.backendServerAddress.String(), path)

	return uh.apiClient.DeleteContent(url)
}
