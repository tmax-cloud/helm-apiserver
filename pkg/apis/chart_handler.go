package apis

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

type Request struct {
	ChartRepositoryName string
}

const (
	indexFileSuffix  = "-index.yaml"
	chartsFileSuffix = "-charts.txt"
)

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

func (hcm *HelmClientManager) InstallChart(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("InstallRelease")
	req := schemas.ChartRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.ReleaseName,
		// ChartName:   path + req.Spec.Path,
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
