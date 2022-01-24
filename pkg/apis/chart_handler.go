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

// 1. main.go 함수 router.HandleFunc(chartPrefix, hcm.GetCharts).Methods("GET") 이 부분 구현은
// repositoryConfig file 읽어서 Helm Repo 이름 list로 받아서
// 해당 repo 이름-index.yaml 파일 정보 다 읽어서 response로 보내야 할 것 같습니다
// repo_handler.go 참고하심 돼요

// 2. main.go 함수 router.HandleFunc(chartPrefix+"/{chart-name}", hcm.GetCharts).Methods("GET") 이 부분 구현은
// 특정 차트의 정보를 GET 하는 요청인데 차트 index 정보 + values.yaml 정보 response로 보내줘야 UI에서 사용자가 수정해서 request로 다시 POST 요청 할 수 있어서
// chart-name 을 path-variable로 받아서 (mux package mux.Vars 함수 이용) 해당 chart의 values.yaml 정보를 보내줘야 할 것 같습니다
// HelmClient.getChart -> chart.Values 활용하면 될듯

// 3. main.go 함수 router.HandleFunc(chartPrefix, hcm.GetCharts).Methods("GET") 이 부분 구현 에서 query parameter가 전달 된 경우
// 아마도 category로 chart 분류하는 경우가 될텐데, 이 경우도 1번에서 읽은 index.yaml 정보 가지고 있는 상태에서
// category 분류해서 해당 되는 차트의 index 정보만 response로 보내줘야 할 것 같습니다
// 이건 아직 chart.yaml 에서 category 필드 추가를 안해 놓은 상태라서 일단 보류~

func (hcm *HelmClientManager) GetCharts(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get Charts")
	w.Header().Set("Content-Type", "application/json")
	req := &schemas.RepoRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

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
