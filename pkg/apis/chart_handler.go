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

	// store repo names into repoNames slice
	var repoNames []string
	for _, repository := range repoList.Repositories {
		repoNames = append(repoNames, repository.Name)
	}

	response := &schemas.ChartResponse{}
	index := &repo.IndexFile{}
	allEntries := make(map[string]repo.ChartVersions)
	repositoryEntries := make(map[string]repo.ChartVersions)
	searchEntries := make(map[string]repo.ChartVersions)

	// read all index.yaml file and save only Entries
	for _, repoName := range repoNames {
		if index, err = readRepoIndex(repoName); err != nil {
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while reading index.yaml file of " + repoName,
			})
			return
		}

		for key, value := range index.Entries {
			allEntries[key] = value
		}
	}

	query, _ := url.ParseQuery(r.URL.RawQuery)
	_, repository := query["repository"]
	_, search := query["search"]

	// in case of query parameter "repository" is requested
	if repository {
		r_index := &repo.IndexFile{}
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

	onlyOneEntries := make(map[string]repo.ChartVersions)
	var chartVersions []*repo.ChartVersion
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

func readRepoIndex(repoName string) (index *repo.IndexFile, err error) {
	index = &repo.IndexFile{}
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
