package apis

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

func (hcm *HelmClientManager) InstallRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("InstallRelease")
	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.Spec.ReleaseName,
		// ChartName:   path + req.Spec.Path,
		ChartName:   req.Spec.Repository,
		Namespace:   req.Namespace,
		ValuesYaml:  req.Values,
		Version:     req.Spec.Version,
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
