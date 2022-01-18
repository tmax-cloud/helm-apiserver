package apis

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

// [TODO] : 구현
func (hcm *HelmClientManager) GetCharts(w http.ResponseWriter, r *http.Request) {

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
