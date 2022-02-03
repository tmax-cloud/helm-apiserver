package apis

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"k8s.io/klog"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
)

const (
	indexFileSuffix  = "-index.yaml"
	chartsFileSuffix = "-charts.txt"
)

// 1. category 분류 query로 할지 UI에서 진행할지 확인 필요 (UI)
// 2. Repo 별로 chart 따로 response 할지 확인 필요 (기획)

func (hcm *HelmClientManager) GetCharts(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get Charts")
	w.Header().Set("Content-Type", "application/json")

	// Read repositoryConfig File which contains repo Info list
	repoList := &schemas.RepositoryFile{}
	repoListFile, err := ioutil.ReadFile(repositoryConfig)
	if err != nil {
		klog.Errorln(err, "failed to get repository list file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
		})
		return
	}

	repoListFileJson, _ := yaml.YAMLToJSON(repoListFile) // Should transform yaml to Json

	if err := json.Unmarshal(repoListFileJson, repoList); err != nil {
		klog.Errorln(err, "failed to unmarshal repo file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while unmarshalling request",
		})
		return
	}

	var repoNames []string
	for _, repository := range repoList.Repositories {
		repoNames = append(repoNames, repository.Name)
	}

	response := &schemas.ChartResponse{}
	index := &repo.IndexFile{}
	responseEntries := make(map[string]repo.ChartVersions)

	// read all index.yaml file and save only Entries
	for _, repoName := range repoNames {
		indexFile, err := ioutil.ReadFile(repositoryCache + "/" + repoName + indexFileSuffix)
		if err != nil {
			klog.Errorln(err, "failed to read index.yaml file of "+repoName)
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while reading index.yaml file of " + repoName,
			})
			return
		}

		indexFileJson, _ := yaml.YAMLToJSON(indexFile) // Should transform yaml to Json

		if err := json.Unmarshal(indexFileJson, index); err != nil {
			klog.Errorln(err, "failed to unmarshal index file")
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while unmarshalling index file",
			})
			return
		}

		for key, value := range index.Entries {
			responseEntries[key] = value
		}
	}

	// in case of {chart-name} is requested
	vars := mux.Vars(r)
	reqChartName, exist := vars["chart-name"]

	onlyOneEntries := make(map[string]repo.ChartVersions)
	var chartVersions []*repo.ChartVersion
	var reqURL string

	if exist {
		index.Entries[reqChartName] = responseEntries[reqChartName]
		for _, chart := range index.Entries[reqChartName] {
			if chart.Name == reqChartName {
				chartVersions = append(chartVersions, chart)
				onlyOneEntries[reqChartName] = chartVersions

				// get reqURL value for value.yaml
				for _, url := range chart.URLs {
					reqURL = url
				}
			}
		}
		index.Entries = onlyOneEntries
		response.IndexFile = *index

		helmChart, _, _ := hcm.getChart(reqURL, &action.ChartPathOptions{})

		if helmChart == nil {
			klog.Errorln(err, "failed to get chart: "+reqURL+" info")
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while getting chart: " + reqURL + " info",
			})
			return
		}
		response.Values = helmChart.Values

		klog.Infoln("Get Charts of " + reqChartName + " is successfully done")
		respond(w, http.StatusOK, response)
		return
	}

	index.Entries = responseEntries // add all repo's the entries
	response.IndexFile = *index

	klog.Infoln("Get Charts is successfully done")
	respond(w, http.StatusOK, response)

}

func (hcm *HelmClientManager) getChart(chartName string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, string, error) {
	chartPath, err := chartPathOptions.LocateChart(chartName, hcm.Hcs.Settings)
	if err != nil {
		return nil, "", err
	}

	helmChart, err := loader.Load(chartPath)
	if err != nil {
		return nil, "", err
	}

	if helmChart.Metadata.Deprecated {
		hcm.Hcs.DebugLog("WARNING: This chart (%q) is deprecated", helmChart.Metadata.Name)
	}

	return helmChart, chartPath, err
}
