package apis

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"

	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

// [TODO] : fail response 도 만들어야 함
func (hcm *HelmClientManager) GetReleases(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("GetRelease")
	w.Header().Set("Content-Type", "application/json")
	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	releases, err := hcm.Hc.ListDeployedReleases()
	if err != nil {
		klog.Errorln(err, "failed to decode request")
	}

	response := &schemas.ReleaseResponse{}
	vars := mux.Vars(r)
	reqReleaseName, exist := vars["release-name"]
	if exist {
		for _, rel := range releases {
			if rel.Name == reqReleaseName {
				response.Release = append(response.Release, *rel)
			}
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			klog.Errorln(err, "failed to encode response")
		}
	}

	for _, rel := range releases {
		response.Release = append(response.Release, *rel)
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		klog.Errorln(err, "failed to encode response")
	}

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
