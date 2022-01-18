package apis

import (
	"encoding/json"
	"net/http"

	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/klog"
)

type ChartRepository struct {
	Name string `json:"name"`
	URL  string `json:"repoURL"`
}

func (hcm *HelmClientManager) AddChartRepo(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("AddChartRepo")
	cr := ChartRepository{}
	if err := json.NewDecoder(r.Body).Decode(&cr); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	// TODO : Private Repository도 지원해줘야 함
	chartRepo := repo.Entry{
		Name: cr.Name,
		URL:  cr.URL,
	}

	if err := hcm.Hc.AddOrUpdateChartRepo(chartRepo); err != nil {
		klog.Errorln(err, "failed to add chart repo")
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode("OK : Add chart repo"); err != nil {
		klog.Errorln(err, "failed to encode response")
	}
}
