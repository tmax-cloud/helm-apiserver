package apis

import (
	"encoding/json"
	"net/http"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"

	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

// [TODO] : 구현
func (hcm *HelmClientManager) GetReleases(w http.ResponseWriter, r *http.Request) {

}

func (hcm *HelmClientManager) UnInstallRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("UnInstallRelease")
	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	releaseSpec := helmclient.ChartSpec{
		ReleaseName: req.ReleaseName,
		Namespace:   req.Namespace,
	}

	// [TODO] : Namespace check는 안하는지?
	if err := hcm.Hc.UninstallReleaseByName(releaseSpec.ReleaseName); err != nil {
		klog.Errorln(err, "failed to uninstall chart")
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(""); err != nil {
		klog.Errorln(err, "failed to encode response")
	}
}
