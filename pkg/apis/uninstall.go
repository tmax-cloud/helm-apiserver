package apis

import (
	"encoding/json"
	"net/http"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"

	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

func (hcm *HelmClientManager) UnInstallRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("UnInstallRelease")
	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.Spec.ReleaseName,
		ChartName:   req.Spec.Repository,
		Namespace:   req.Namespace,
		ValuesYaml:  req.Values,
		Version:     req.Spec.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	if err := hcm.Hc.UninstallRelease(&chartSpec); err != nil {
		klog.Errorln(err, "failed to uninstall chart")
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(""); err != nil {
		klog.Errorln(err, "failed to encode response")
	}
}
