package apis

import (
	"context"
	"encoding/json"
	"net/http"

	"strings"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"

	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

// Get all-namespace 구현
func (hcm *HelmClientManager) GetAllReleases(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get All Releases")
	w.Header().Set("Content-Type", "application/json")

	releases, err := hcm.listDeployedReleases()
	if err != nil {
		klog.Errorln(err, "failed to get helm release list")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while getting helm release list",
		})
		return
	}

	var customReleases []*schemas.Release
	for _, rel := range releases {
		customRelease := &schemas.Release{}
		copier.Copy(customRelease, rel)
		customReleases = append(customReleases, customRelease)
	}

	for _, rel := range customReleases {
		if err := setObjectsInfo(rel); err != nil {
			klog.Errorln(err, "Error occurs while setting Obejcts field")
		}
	}

	response := &schemas.ReleaseResponse{}

	for _, rel := range customReleases {
		response.Release = append(response.Release, *rel)
	}

	respond(w, http.StatusOK, response)
}

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

	var customReleases []*schemas.Release
	for _, rel := range releases {
		customRelease := &schemas.Release{}
		copier.Copy(customRelease, rel)
		customReleases = append(customReleases, customRelease)
	}

	for _, rel := range customReleases {
		if err := setObjectsInfo(rel); err != nil {
			klog.Errorln(err, "Error occurs while setting Obejcts field")
		}
	}

	response := &schemas.ReleaseResponse{}
	reqReleaseName, exist := vars["release-name"]
	if exist {
		for _, rel := range customReleases {
			if rel.Name == reqReleaseName {
				response.Release = append(response.Release, *rel)
			}
		}

		respond(w, http.StatusOK, response)
		return
	}

	for _, rel := range customReleases {
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
	hcm.SetClientNS(namespace)

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
	hcm.SetClientNS(namespace)
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

// Response에 생성된 object kind 및 name 추가 하기 위함
func setObjectsInfo(rel *schemas.Release) error {
	var temp []string
	// var objects []*runtime.RawExtension // 일부 releases에서 unstr로 변경 안되는 버그
	var object map[string]interface{}

	splits1 := strings.Split(rel.Manifest, "---")

	for _, spl := range splits1 {
		splits2 := strings.Split(spl, ".yaml")
		temp = append(temp, splits2...)
	}

	objMap := make(map[string]string)
	for _, t := range temp {
		if strings.Contains(t, "apiVersion") {
			trans := []byte(strings.TrimSpace(t))
			raw, _ := yaml.YAMLToJSON(trans)
			json.Unmarshal(raw, &object)
			objMap[object["kind"].(string)] = object["metadata"].(map[string]interface{})["name"].(string)
		}
	}

	// 일부 release에서 unstr로변경 안되는 버그 있음 (체크 필요) - 우선은 위 로직으로 진행
	// for _, t := range temp {
	// 	if strings.Contains(t, "apiVersion") {
	// 		trans := []byte(strings.TrimSpace(t))
	// 		raw, _ := yaml.YAMLToJSON(trans) // needed to transform unstr type

	// 		obj := &runtime.RawExtension{
	// 			Raw: raw,
	// 		}
	// 		objects = append(objects, obj)
	// 	}
	// }

	// objMap := make(map[string]string)
	// for _, obj := range objects {
	// 	unstr, err := BytesToUnstructuredObject(obj)
	// 	if err != nil {
	// 		klog.Errorln(err, "failed to transform to unstr type")
	// 		return err
	// 	}
	// 	objMap[unstr.GetKind()] = unstr.GetName()
	// }
	rel.Objects = objMap
	return nil
}

// func BytesToUnstructuredObject(obj *runtime.RawExtension) (*unstructured.Unstructured, error) {
// 	var in runtime.Object
// 	var scope conversion.Scope // While not actually used within the function, need to pass in
// 	if err := runtime.Convert_runtime_RawExtension_To_runtime_Object(obj, &in, scope); err != nil {
// 		return nil, err
// 	}

// 	unstrObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(in)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &unstructured.Unstructured{Object: unstrObj}, nil
// }

// all-namespaces 조회를 위한 함수
func (hcm *HelmClientManager) listDeployedReleases() ([]*release.Release, error) {
	listClient := action.NewList(hcm.Hcs.ActionConfig)

	listClient.StateMask = action.ListDeployed
	listClient.AllNamespaces = true // 추가됨
	listClient.All = true           // 추가됨

	return listClient.Run()
}
