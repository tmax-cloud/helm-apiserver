package apis

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"

	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

// [TODO]
// 1. all-namespace 구현
func (hcm *HelmClientManager) GetReleases(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("GetRelease")
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	namespace := vars["ns-name"]
	hcm.SetClientNS(namespace)

	releases, err := hcm.Hci.ListDeployedReleases()
	if err != nil {
		klog.Errorln(err, "failed to get helm release list")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while getting helm release list",
		})
		return
	}

	response := &schemas.ReleaseResponse{}

	reqReleaseName, exist := vars["release-name"]
	if exist {
		for _, rel := range releases {
			if rel.Name == reqReleaseName {
				response.Release = append(response.Release, *rel)
			}
		}

		respond(w, http.StatusOK, response)
		return
	}

	for _, rel := range releases {
		response.Release = append(response.Release, *rel)
	}

	respond(w, http.StatusOK, response)
}

func (hcm *HelmClientManager) InstallRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("InstallRelease")
	w.Header().Set("Content-Type", "application/json")
	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

	vars := mux.Vars(r)
	namespace := vars["ns-name"]

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.ReleaseName,
		ChartName:   req.PackageURL,
		Namespace:   namespace,
		ValuesYaml:  req.Values,
		Version:     req.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	// [TODO] ChartIsInstalled check 해야 하나?
	if _, err := hcm.Hci.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
		klog.Errorln(err, "failed to install release")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while installing helm release",
		})
		return
	}

	respond(w, http.StatusOK, req.ReleaseName+" release is successfully installed")
}

// 일단 안씀
func (hcm *HelmClientManager) RollbackRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Rollback Release")
	w.Header().Set("Content-Type", "application/json")
	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

	vars := mux.Vars(r)
	namespace := vars["ns-name"]
	reqReleaseName := vars["release-name"]

	chartSpec := helmclient.ChartSpec{
		ReleaseName: reqReleaseName,
		ChartName:   req.PackageURL,
		Namespace:   namespace,
		ValuesYaml:  req.Values,
		Version:     req.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	if err := hcm.Hci.RollbackRelease(&chartSpec, 0); err != nil {
		klog.Errorln(err, "failed to rollback chart")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while rollback helm release",
		})
	}

	respond(w, http.StatusOK, reqReleaseName+" release is successfully rollbacked")
}

func (hcm *HelmClientManager) UnInstallRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("UnInstallRelease")
	w.Header().Set("Content-Type", "application/json")
	// req := schemas.ReleaseRequest{}
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	klog.Errorln(err, "failed to decode request")
	// 	respond(w, http.StatusBadRequest, &schemas.Error{
	// 		Error:       err.Error(),
	// 		Description: "Error occurs while decoding request",
	// 	})
	// 	return
	// }

	vars := mux.Vars(r)
	namespace := vars["ns-name"]
	reqReleaseName := vars["release-name"]

	hcm.SetClientNS(namespace)
	if err := hcm.Hci.UninstallReleaseByName(reqReleaseName); err != nil {
		klog.Errorln(err, "failed to uninstall chart")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while uninstalling helm release",
		})
		return
	}

	respond(w, http.StatusOK, reqReleaseName+" release is successfully uninstalled")
}

func respond(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		klog.Errorln(err, "Error occurs while encoding response body")
	}
}
