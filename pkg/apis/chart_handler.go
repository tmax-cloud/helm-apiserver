package apis

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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

// type ChartHandler struct {
// 	router *mux.Router
// 	Hcm    *HelmClientManager
// }

// func (c *ChartHandler) Init(router *mux.Router) {
// 	c.router = router
// 	c.router.HandleFunc(chartPrefix, c.Hcm.GetCharts).Methods("GET")                 // 설치 가능한 chart list 반환
// 	c.router.HandleFunc(chartPrefix+"/{chart-name}", c.Hcm.GetCharts).Methods("GET") // (query : category 분류된 chart list반환 / path-varaible : 특정 chart data + value.yaml 반환)
// 	http.Handle("/", c.router)
// }

func (hcm *HelmClientManager) GetCharts(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get Charts")
	setResponseHeader(w)

	// sync the latest charts
	// if err := hcm.updateChartRepo(); err != nil {
	// 	respond(w, http.StatusBadRequest, &schemas.Error{
	// 		Error:       err.Error(),
	// 		Description: "Error occurs while sync the latest charts",
	// 	})
	// 	return
	// }

	// Read repositoryConfig File which contains repo Info list
	repoList, err := readRepoList()
	if err != nil {
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
		})
		return
	}

	repoInfos := make(map[string]string)
	// store repo names into repoNames slice
	for _, repoInfo := range repoList.Repositories {
		repoInfos[repoInfo.Name] = repoInfo.Url
	}

	response := &schemas.ChartResponse{}
	index := &schemas.IndexFile{}
	allEntries := make(map[string]schemas.ChartVersions)
	repositoryEntries := make(map[string]schemas.ChartVersions)
	searchEntries := make(map[string]schemas.ChartVersions)

	// read all index.yaml file and save only Entries
	for repoName, repoUrl := range repoInfos {
		if index, err = readRepoIndex(repoName); err != nil {
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while reading index.yaml file of " + repoName,
			})
			return
		}

		// add repo info
		for key, value := range index.Entries {
			for _, chart := range value {
				chart.Repo.Name = repoName
				chart.Repo.Url = repoUrl
			}
			allEntries[key] = value
		}
	}

	query, _ := url.ParseQuery(r.URL.RawQuery)
	_, repository := query["repository"]
	_, search := query["search"]

	// in case of query parameter "repository" is requested
	if repository {
		r_index := &schemas.IndexFile{}
		repoName := query.Get("repository")
		if r_index, err = readRepoIndex(repoName); err != nil {
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while reading index.yaml file of " + repoName,
			})
			return
		}

		for key, value := range r_index.Entries {
			repositoryEntries[key] = value
		}
	}

	// in case of query parameter "search" is requested
	if search {
		var keywords []string
		searcher := query.Get("search")

		if repository {
			for key, value := range repositoryEntries {
				if strings.Contains(key, searcher) {
					searchEntries[key] = value
					continue // go to next key
				}

				for _, chart := range value {
					keywords = chart.Keywords
				}

				for _, keyword := range keywords {
					if strings.Contains(keyword, searcher) {
						searchEntries[key] = value
						break // break present for loop
					}
				}
			}
		} else {
			for key, value := range allEntries {
				if strings.Contains(key, searcher) {
					searchEntries[key] = value
					continue // go to next key
				}

				for _, chart := range value {
					keywords = chart.Keywords
				}

				for _, keyword := range keywords {
					if strings.Contains(keyword, searcher) {
						searchEntries[key] = value
						break // break present for loop
					}
				}
			}
		}
	}

	// in case of {chart-name} is requested
	vars := mux.Vars(r)
	reqChartName, exist := vars["chart-name"]

	onlyOneEntries := make(map[string]schemas.ChartVersions)
	var chartVersions []*schemas.ChartVersion
	var reqURL string

	if exist {
		for _, chart := range allEntries[reqChartName] {
			if chart.Name == reqChartName {
				chartVersions = append(chartVersions, chart)
				onlyOneEntries[reqChartName] = chartVersions

				// get reqURL value for values.yaml
				for _, url := range chart.URLs {
					reqURL = url
				}
			}
		}
		index.Entries = onlyOneEntries
		response.IndexFile = *index

		helmChart, _, err := hcm.getChart(reqURL, &action.ChartPathOptions{})

		if helmChart == nil {
			klog.Errorln(err, "failed to get chart: "+reqChartName+" info")
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while getting chart: " + reqChartName + " info",
			})
			return
		}
		response.Values = helmChart.Values

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
		index.Entries = allEntries // set all repo's the entries
		response.IndexFile = *index
	}

	klog.Infoln("Get Charts is successfully done")
	respond(w, http.StatusOK, response)

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
		klog.Errorln(err, "failed to get repository list file")
		return nil, err
	}

	repoListFileJson, _ := yaml.YAMLToJSON(repoListFile) // Should transform yaml to Json

	if err = json.Unmarshal(repoListFileJson, repoList); err != nil {
		klog.Errorln(err, "failed to unmarshal repo file")
		return nil, err
	}

	return repoList, nil
}

// [TODO] : Private repo별로 TLS client 세팅 해줘야함...
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

		if err := hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
			klog.Errorln(err, "failed to update chart repo")
			return err
		}
	}
	return nil
}
