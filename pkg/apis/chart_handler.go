package apis

import (
	"bufio"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"

	helmclient "github.com/mittwald/go-helm-client"
)

const (
	indexFileSuffix  = "-index.yaml"
	chartsFileSuffix = "-charts.txt"
)

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

func (hcm *HelmClientManager) InstallChart(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("InstallRelease")
	req := schemas.ChartRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.ReleaseName,
		ChartName:   req.PackageURL,
		Namespace:   req.Namespace,
		ValuesYaml:  req.Values,
		Version:     req.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	if _, err := hcm.Hc.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
		klog.Errorln(err, "failed to install chart")
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(""); err != nil {
		klog.Errorln(err, "failed to encode response")
	}
}

// 일단 안씀
func (hcm *HelmClientManager) RollbackRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("RollbackRelease")
	req := schemas.ChartRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.ReleaseName,
		ChartName:   req.PackageURL,
		Namespace:   req.Namespace,
		ValuesYaml:  req.Values,
		Version:     req.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	if err := hcm.Hc.RollbackRelease(&chartSpec, 0); err != nil {
		klog.Errorln(err, "failed to rollback chart")
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(""); err != nil {
		klog.Errorln(err, "failed to encode response")
	}
}
