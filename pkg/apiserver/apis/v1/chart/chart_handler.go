package chart

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/klog"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/tmax-cloud/helm-apiserver/internal/utils"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
)

func (c *ChartHandler) chartHandler(w http.ResponseWriter, req *http.Request) {
	c.GetCharts(w, req)
	klog.Info(req.Header)
}

func (ch *ChartHandler) GetCharts(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get Charts")
	utils.SetResponseHeader(w)

	if ch.Index == nil {
		utils.Respond(w, http.StatusOK, &schemas.Error{
			Error:       "No helm repository is added",
			Description: "you need to add at least one helm repository",
		})
		return
	}

	startTime := time.Now() // for checking response time

	response := &schemas.ChartResponse{}
	index := &schemas.IndexFile{}
	repositoryEntries := make(map[string]schemas.ChartVersions)
	searchEntries := make(map[string]schemas.ChartVersions)

	query, _ := url.ParseQuery(r.URL.RawQuery)
	_, repository := query["repository"]
	_, search := query["search"]

	// in case of query parameter "repository" is requested
	if repository {
		repoName := query.Get("repository")
		for key, charts := range ch.SingleChartEntries {
			for _, chart := range charts {
				if chart.Repo.Name == repoName {
					repositoryEntries[key] = charts
				}
			}
		}
	}

	// in case of query parameter "search" is requested
	if search {
		var keywords []string
		searcher := query.Get("search")

		if repository {
			for key, charts := range repositoryEntries {
				if strings.Contains(key, searcher) {
					searchEntries[key] = charts
					continue // go to next key
				}

				for _, chart := range charts {
					keywords = chart.Keywords
				}

				for _, keyword := range keywords {
					if strings.Contains(keyword, searcher) {
						searchEntries[key] = charts
						break // break present for loop
					}
				}
			}
		} else {
			for key, charts := range ch.SingleChartEntries {
				if strings.Contains(key, searcher) {
					searchEntries[key] = charts
					continue // go to next key
				}

				for _, chart := range charts {
					keywords = chart.Keywords
				}

				for _, keyword := range keywords {
					if strings.Contains(keyword, searcher) {
						searchEntries[key] = charts
						break // break present for loop
					}
				}
			}
		}
	}

	vars := mux.Vars(r)
	reqChartName, existChart := vars["chart-name"]
	reqVersion, existVersion := vars["version"]

	onlyOneEntries := make(map[string]schemas.ChartVersions)
	// 특정 차트에 대한 응답
	if existChart {
		// 특정 버전에 대한 응답
		if existVersion {
			var selectedChart *schemas.ChartVersion
			for _, chart := range ch.Index.Entries[reqChartName] {
				if chart.Version == reqVersion {
					selectedChart = chart
				}
			}
			// error check 필요
			charts, values, _ := getChartInfo(ch, selectedChart)
			onlyOneEntries[reqChartName] = charts
			index.Entries = onlyOneEntries
			response.IndexFile = *index
			response.Values = values
		} else {
			for _, chart := range ch.SingleChartEntries[reqChartName] {
				charts, values, _ := getChartInfo(ch, chart)
				onlyOneEntries[reqChartName] = charts
				index.Entries = onlyOneEntries
				response.IndexFile = *index
				response.Values = values
			}

			// 특정 차트 선택 시 사용 가능한 version 추가
			var versions []string
			for _, chart := range ch.Index.Entries[reqChartName] {
				versions = append(versions, chart.Version)
			}
			response.Versions = versions
		}

		checkTime2 := time.Since(startTime)
		klog.Info("check1 Duration: ", checkTime2)

		klog.Infoln("Get Charts of " + reqChartName + " is successfully done")
		utils.Respond(w, http.StatusOK, response)
		return
	}

	// set response following switch cases of query params
	switch {
	case repository && !search:
		index.Entries = repositoryEntries // set requested repo's the entries
		response.IndexFile = *index
	case search:
		index.Entries = searchEntries // set requested search's the entries
		response.IndexFile = *index
	default:
		index.Entries = ch.SingleChartEntries // set all repo's the entries
		response.IndexFile = *index
	}

	klog.Infoln("Get Charts is successfully done")
	utils.Respond(w, http.StatusOK, response)

	elapsedTime := time.Since(startTime)
	klog.Info("Total Duration: ", elapsedTime)

}

func getChartInfo(ch *ChartHandler, chart *schemas.ChartVersion) (schemas.ChartVersions, map[string]interface{}, error) {
	var chartVersions []*schemas.ChartVersion
	var reqURL string
	var chartPath string

	chartPaths := []string{}
	// get reqURL value for values.yaml
	for _, url := range chart.URLs {
		reqURL = url
		chartPath = url
	}

	// default index 파일은 필요함
	if !strings.Contains(chartPath, chart.Repo.Url) {
		chartPath = chart.Repo.Url + "/" + reqURL
		chartPaths = append(chartPaths, chartPath)
		chart.URLs = chartPaths
	}

	chartVersions = append(chartVersions, chart)

	// getChart 후 /tmp/.helmcache/ 에 파일 저장됨
	helmChart, filePath, err := ch.getChart(chartPath, &action.ChartPathOptions{
		InsecureSkipTLSverify: true,
	})
	defer os.Remove(filePath)

	// klog.Info("File NAME")
	// for _, file := range helmChart.Raw {
	// 	if file.Name == "values.yaml" {
	// 		klog.Info(string(file.Data))
	// 	}
	// }
	// klog.Info(string(helmChart.Schema))

	if helmChart == nil {
		klog.Errorln(err, "failed to get chart: "+chart.Name+" info")
		return nil, nil, err
	}

	return chartVersions, helmChart.Values, nil

}

// this function is for get values info
func (ch *ChartHandler) getChart(chartName string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, string, error) {
	chartPath, err := chartPathOptions.LocateChart(chartName, ch.hcm.Hcs.Settings)
	if err != nil {
		return nil, "", err
	}

	helmChart, err := loader.Load(chartPath)
	if err != nil {
		return nil, "", err
	}

	if helmChart.Metadata.Deprecated {
		ch.hcm.Hcs.DebugLog("WARNING: This chart (%q) is deprecated", helmChart.Metadata.Name)
	}

	return helmChart, chartPath, err

}
