package apis

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"k8s.io/klog"
	"sigs.k8s.io/yaml"
	helmclient "github.com/mittwald/go-helm-client"
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

// 설치 가능한 chart list 반환
func (hcm *HelmClientManager) GetCharts(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("GetCharts")
	repositoriesFile, err := ioutil.ReadFile(repositoryConfig)
	if err != nil {
		klog.Errorln(err, "failed to read repo List file")
		return
	}
	reposJsonFile, _ := yaml.YAMLToJSON(repositoriesFile)

	reposFile := &schemas.RepositoryFile{}
	if err := json.Unmarshal(reposJsonFile, reposFile); err != nil {
		klog.Errorln(err, "failed to unmarshal repo file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while unmarshalling request",
		})
		return
	}

	response := &schemas.ChartsResponse{}
	for _, repo := range reposFile.Repositories {
		indexFile, err := ioutil.ReadFile(repositoryCache + "/" + repo.Name + indexFileSuffix)
		if err != nil {
			klog.Errorln(err, "failed to read index file of "+repo.Name)
			return
		}

		idxJsonFile, _ := yaml.YAMLToJSON(indexFile)

		idxFile := &schemas.IndexFile{}
		if err := json.Unmarshal(idxJsonFile, idxFile); err != nil {
			klog.Errorln(err, "failed to unmarshal index file of "+repo.Name)
			return
		}

		entries := idxFile.Entries
		chartNames, _ := getChartNames(repo.Name)
		for _, chartName := range chartNames {
			chartinfos := entries[chartName]
			for _, chartinfo := range chartinfos {
				response.ChartInfos = append(response.ChartInfos, *chartinfo)
			}
		}
	}

	respond(w, http.StatusOK, response)
}

func getChartNames(ChartRepoName string) ([]string, error) {
	// open the file
	file, err := os.Open(repositoryCache + "/" + ChartRepoName + chartsFileSuffix)

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
	}

	file.Close()
	return chartList, nil
}

