package apis

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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

type ChartHandler struct {
	hcm                *HelmClientManager
	Index              *schemas.IndexFile
	SingleChartEntries map[string]schemas.ChartVersions
	// col                *mongo.Collection
}

func NewChartHandler(hcmanager *HelmClientManager) *ChartHandler {
	index := getIndex()
	singleChartEntries := getSingleChart(index)

	klog.Info("Setting ChartHandler is done")
	return &ChartHandler{
		hcm:                hcmanager,
		Index:              index,
		SingleChartEntries: singleChartEntries,
	}

}

func (ch *ChartHandler) UpdateChartHandler() {
	ch.Index = getIndex()
	ch.SingleChartEntries = getSingleChart(ch.Index)
	klog.Info("Updating ChartHandler is done")
}

func getIndex() *schemas.IndexFile {
	// Read repositoryConfig File which contains repo Info list
	repoList, err := readRepoList()
	if err != nil {
		klog.Errorln(err, "failed to save index file")
		return nil
	}

	repoInfos := make(map[string]string)
	// store repo names
	for _, repoInfo := range repoList.Repositories {
		repoInfos[repoInfo.Name] = repoInfo.Url
	}

	index := &schemas.IndexFile{}
	allEntries := make(map[string]schemas.ChartVersions)
	// col := db.GetMongoDBConnetion() // #######테스트########

	// read all index.yaml file and save only Entries
	for repoName, repoUrl := range repoInfos {
		if index, err = readRepoIndex(repoName); err != nil {
			klog.Errorln(err, "failed to read index file")
		}

		// add repo info
		for key, charts := range index.Entries {
			for _, chart := range charts {
				chart.Repo.Name = repoName
				chart.Repo.Url = repoUrl
				// _, err := db.InsertDoc(col, chart) // #######테스트########
				// klog.Info("insert done!")
				// if err != nil {
				// 	klog.Error(err)
				// }
			}
			allEntries[repoName+"_"+key] = charts // 중복 chart name 가능하도록 repo name과 결합
		}
	}

	// filter := bson.D{{}}
	// var test []schemas.ChartVersion
	// test, _ = db.FindDoc(col, filter, filter)
	// for _, ch := range test {
	// 	klog.Info(ch.Name)
	// }

	index.Entries = allEntries
	klog.Info("saving index file is done")
	return index
}

func getSingleChart(index *schemas.IndexFile) map[string]schemas.ChartVersions {
	if index == nil {
		return nil
	}

	singleChartEntries := make(map[string]schemas.ChartVersions)
	for key, charts := range index.Entries {
		var oneChart []*schemas.ChartVersion
		oneChart = append(oneChart, charts[0])
		singleChartEntries[key] = oneChart
	}
	return singleChartEntries
}

func (ch *ChartHandler) GetCharts(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get Charts")
	setResponseHeader(w)

	if ch.Index == nil {
		respond(w, http.StatusOK, &schemas.Error{
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
			charts, values, _ := getChartInfo(ch.hcm, selectedChart)
			onlyOneEntries[reqChartName] = charts
			index.Entries = onlyOneEntries
			response.IndexFile = *index
			response.Values = values
		} else {
			for _, chart := range ch.SingleChartEntries[reqChartName] {
				charts, values, _ := getChartInfo(ch.hcm, chart)
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
		respond(w, http.StatusOK, response)
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
	respond(w, http.StatusOK, response)

	elapsedTime := time.Since(startTime)
	klog.Info("Total Duration: ", elapsedTime)

}

func getChartInfo(hcm *HelmClientManager, chart *schemas.ChartVersion) (schemas.ChartVersions, map[string]interface{}, error) {
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
	helmChart, filePath, err := hcm.getChart(chartPath, &action.ChartPathOptions{
		InsecureSkipTLSverify: true,
	})

	if helmChart == nil {
		klog.Errorln(err, "failed to get chart: "+chart.Name+" info")
		return nil, nil, err
	}
	defer os.Remove(filePath)

	return chartVersions, helmChart.Values, nil

}

// this function is for get values info
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

func readRepoIndex(repoName string) (index *schemas.IndexFile, err error) {
	index = &schemas.IndexFile{}

	indexFile, err := ioutil.ReadFile(repositoryCache + "/" + repoName + indexFileSuffix)
	if err != nil {
		klog.Errorln(err, "failed to read index.yaml file of "+repoName)
		return nil, err
	}

	indexFileJson, _ := yaml.YAMLToJSON(indexFile) // Should transform yaml to Json

	if err := json.Unmarshal(indexFileJson, index); err != nil {
		klog.Errorln(err, "failed to unmarshal index file")
		return nil, err
	}

	return index, nil
}

func readRepoList() (repoList *schemas.RepositoryFile, err error) {
	repoList = &schemas.RepositoryFile{}
	repoListFile, err := ioutil.ReadFile(repositoryConfig)
	if err != nil {
		klog.Errorln(err, "failed to get repository list")
		return nil, err
	}

	repoListFileJson, _ := yaml.YAMLToJSON(repoListFile) // Should transform yaml to Json

	if err = json.Unmarshal(repoListFileJson, repoList); err != nil {
		klog.Errorln(err, "failed to unmarshal repo file")
		return nil, err
	}

	return repoList, nil
}

// [TODO] : 전체 말고 차트 불러오는 repo만 update로 변경?
func (hcm *HelmClientManager) updateChartRepo() error {
	klog.Infoln("Sync the latest chart info")

	// Read repositoryConfig File which contains repo Info list
	repoList, err := readRepoList()
	if err != nil {
		klog.Errorln(err, "failed to get repoList while sync latest chart info")
		return err
	}

	chartRepo := repo.Entry{}
	for _, repo := range repoList.Repositories {
		chartRepo.Name = repo.Name
		chartRepo.URL = repo.Url
		// chartRepo.CAFile = repo.CaFile
		// chartRepo.KeyFile = repo.KeyFile
		// chartRepo.CertFile = repo.CertFile
		chartRepo.InsecureSkipTLSverify = true

		if err := hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
			klog.Errorln(err, "failed to update chart repo")
			return err
		}
	}
	return nil
}

// func (hcm *HelmClientManager) GetCharts(w http.ResponseWriter, r *http.Request) {
// 	klog.Infoln("Get Charts")
// 	setResponseHeader(w)
// 	startTime := time.Now()

// 	// // sync the latest charts
// 	// if err := hcm.updateChartRepo(); err != nil {
// 	// 	respond(w, http.StatusBadRequest, &schemas.Error{
// 	// 		Error:       err.Error(),
// 	// 		Description: "Error occurs while sync the latest charts",
// 	// 	})
// 	// 	return
// 	// }

// 	// Read repositoryConfig File which contains repo Info list
// 	repoList, err := readRepoList()
// 	if err != nil {
// 		respond(w, http.StatusBadRequest, &schemas.Error{
// 			Error:       err.Error(),
// 			Description: "Error occurs while reading repository list file",
// 		})
// 		return
// 	}

// 	repoInfos := make(map[string]string)
// 	// store repo names into repoNames slice
// 	for _, repoInfo := range repoList.Repositories {
// 		repoInfos[repoInfo.Name] = repoInfo.Url
// 	}

// 	response := &schemas.ChartResponse{}
// 	index := &schemas.IndexFile{}
// 	allEntries := make(map[string]schemas.ChartVersions)
// 	repositoryEntries := make(map[string]schemas.ChartVersions)
// 	searchEntries := make(map[string]schemas.ChartVersions)

// 	// read all index.yaml file and save only Entries
// 	for repoName, repoUrl := range repoInfos {
// 		if index, err = readRepoIndex(repoName); err != nil {
// 			respond(w, http.StatusBadRequest, &schemas.Error{
// 				Error:       err.Error(),
// 				Description: "Error occurs while reading index.yaml file of " + repoName,
// 			})
// 			return
// 		}

// 		// add repo info
// 		for key, value := range index.Entries {
// 			for _, chart := range value {
// 				chart.Repo.Name = repoName
// 				chart.Repo.Url = repoUrl
// 			}
// 			allEntries[repoName+key] = value
// 		}
// 	}

// 	query, _ := url.ParseQuery(r.URL.RawQuery)
// 	_, repository := query["repository"]
// 	_, search := query["search"]

// 	// in case of query parameter "repository" is requested
// 	if repository {
// 		r_index := &schemas.IndexFile{}
// 		repoName := query.Get("repository")
// 		if r_index, err = readRepoIndex(repoName); err != nil {
// 			respond(w, http.StatusBadRequest, &schemas.Error{
// 				Error:       err.Error(),
// 				Description: "Error occurs while reading index.yaml file of " + repoName,
// 			})
// 			return
// 		}

// 		for key, value := range r_index.Entries {
// 			repositoryEntries[key] = value
// 		}
// 	}

// 	// in case of query parameter "search" is requested
// 	if search {
// 		var keywords []string
// 		searcher := query.Get("search")

// 		if repository {
// 			for key, value := range repositoryEntries {
// 				if strings.Contains(key, searcher) {
// 					searchEntries[key] = value
// 					continue // go to next key
// 				}

// 				for _, chart := range value {
// 					keywords = chart.Keywords
// 				}

// 				for _, keyword := range keywords {
// 					if strings.Contains(keyword, searcher) {
// 						searchEntries[key] = value
// 						break // break present for loop
// 					}
// 				}
// 			}
// 		} else {
// 			for key, value := range allEntries {
// 				if strings.Contains(key, searcher) {
// 					searchEntries[key] = value
// 					continue // go to next key
// 				}

// 				for _, chart := range value {
// 					keywords = chart.Keywords
// 				}

// 				for _, keyword := range keywords {
// 					if strings.Contains(keyword, searcher) {
// 						searchEntries[key] = value
// 						break // break present for loop
// 					}
// 				}
// 			}
// 		}
// 	}

// 	// in case of {chart-name} is requested
// 	vars := mux.Vars(r)
// 	reqChartName, exist := vars["chart-name"]

// 	onlyOneEntries := make(map[string]schemas.ChartVersions)
// 	var chartVersions []*schemas.ChartVersion
// 	var reqURL string
// 	var chartPath string

// 	// [TODO] 특정 chart GET 할 경우 url 변경해주기
// 	if exist {
// 		for _, chart := range allEntries[reqChartName] {
// 			if chart.Name == reqChartName {

// 				chartPaths := []string{}
// 				// get reqURL value for values.yaml
// 				for _, url := range chart.URLs {
// 					reqURL = url
// 					chartPath = url
// 				}

// 				// default index 파일은 필요함
// 				if !strings.Contains(chartPath, chart.Repo.Url) {
// 					chartPath = chart.Repo.Url + "/" + reqURL
// 					chartPaths = append(chartPaths, chartPath)
// 					chart.URLs = chartPaths
// 				}

// 				chartVersions = append(chartVersions, chart)
// 				onlyOneEntries[reqChartName] = chartVersions

// 			}
// 		}
// 		index.Entries = onlyOneEntries
// 		response.IndexFile = *index

// 		// getChart 후 /tmp/.helmcache/ 에 파일 저장됨
// 		helmChart, filePath, err := hcm.getChart(chartPath, &action.ChartPathOptions{
// 			InsecureSkipTLSverify: true,
// 		})

// 		defer os.Remove(filePath)

// 		if helmChart == nil {
// 			klog.Errorln(err, "failed to get chart: "+reqChartName+" info")
// 			respond(w, http.StatusBadRequest, &schemas.Error{
// 				Error:       err.Error(),
// 				Description: "Error occurs while getting chart: " + reqChartName + " info",
// 			})
// 			return
// 		}
// 		response.Values = helmChart.Values

// 		klog.Infoln("Get Charts of " + reqChartName + " is successfully done")
// 		respond(w, http.StatusOK, response)
// 		return
// 	}

// 	// set response following switch cases of query params
// 	switch {
// 	case repository && !search:
// 		index.Entries = repositoryEntries // set requested repo's the entries
// 		response.IndexFile = *index
// 	case search:
// 		index.Entries = searchEntries // set requested search's the entries
// 		response.IndexFile = *index
// 	default:
// 		index.Entries = allEntries // set all repo's the entries
// 		response.IndexFile = *index
// 	}

// 	klog.Infoln("Get Charts is successfully done")
// 	respond(w, http.StatusOK, response)
// 	elapsedTime := time.Since(startTime)
// 	klog.Info("Time Duration: ", elapsedTime)

// }
