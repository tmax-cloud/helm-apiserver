package apis

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"k8s.io/klog"

	"helm.sh/helm/v3/pkg/chart"
	// "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/repo"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
)

// Request는 schemas package에서 일괄 관리하는게 좋을 것 같습니다
type Request struct {
	ChartRepositoryName string
}

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

// 설치 가능한 chart list 반환
// AddChartRepo로 helm-charts 레포를 등록했다고 가정
// helmcache에서 {chart-repo-name}-index.yaml
// helmcache에서 {chart-repo-name}-charts.txt
func (hcm *HelmClientManager) GetCharts(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("GetCharts")
	//req := schemas.ChartRequest{}
	req := Request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	charts, err := getChartList(req.ChartRepositoryName)
	if err != nil {
		klog.Errorln(err, "failed to get chart list")
		return
	}

	for i, s := range charts {
		fmt.Println(i, s)
	}
}

func getChartList(ChartRepoName string) ([]string, error) {
	// open the file
	file, err := os.Open(repositoryCache + "/" + ChartRepoName + chartsFileSuffix)

	defer file.Close()

	//handle errors while opening
	if err != nil {
		klog.Errorln(err, "Error when opening file")
		return nil, err
	}

	fileScanner := bufio.NewScanner(file)

	var chartList []string

	// read line by line
	for fileScanner.Scan() {
		line := fileScanner.Text()
		chartList = append(chartList, line)
		fmt.Println(line)
	}

	return chartList, nil
}

func (hcm *HelmClientManager) GetCharts2(w http.ResponseWriter, r *http.Request) {
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
	for _, repository := range repoList.Repositories { // check 필요
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

	helmChart := &chart.Chart{}
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

		// helmChart, _ = loader.LoadFile(reqURL)
		// klog.Infoln(helmChart) // for test -> nil 이라 response에 추가할때 nil pointer 에러남
		/////////////////// 이부분 ///////////////////
		// err 처리 추가

		byteChart, err := hcm.Hc.TemplateChart(&helmclient.ChartSpec{
			ReleaseName: "temp",
			ChartName:   reqURL,
			Namespace:   "default",
		})
		if err != nil {
			klog.Infoln(err)
		}
		klog.Infoln(string(byteChart))                 // for test
		helmChartJson, _ := yaml.YAMLToJSON(byteChart) // Should transform yaml to Json

		klog.Infoln("------------------------")
		klog.Infoln(helmChartJson)

		if err := json.Unmarshal(helmChartJson, helmChart); err != nil {
			klog.Errorln(err, "failed to unmarshal rendered chart")
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while unmarshalling rendered chart",
			})
			return
		}
		klog.Infoln("------------------------")
		klog.Infoln(helmChart)

		response.Values = helmChart.Values
		/////////////////////////////////////////////////

		klog.Infoln("Get Charts of " + reqChartName + " is successfully done")
		respond(w, http.StatusOK, response)
		return
	}

	index.Entries = responseEntries // add all repo's the entries
	response.IndexFile = *index

	klog.Infoln("Get Charts is successfully done")
	respond(w, http.StatusOK, response)

}
