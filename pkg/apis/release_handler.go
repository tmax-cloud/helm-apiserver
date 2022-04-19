package apis

import (
	"context"
	"encoding/json"
	"net/http"

	"strings"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"

	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

// type ReleaseHandler struct {
// 	router *mux.Router
// 	Hcm    *HelmClientManager
// }

// func (r *ReleaseHandler) Init(router *mux.Router) {
// 	r.router = router
// 	r.router.HandleFunc(allNamespaces+releasePrefix, r.Hcm.GetAllReleases).Methods("GET") // all-namespaces 릴리즈 반환
// 	r.router.HandleFunc(nsPrefix+releasePrefix, r.Hcm.GetReleases).Methods("GET")         // 설치된 release list 반환 (path-variable : 특정 release 정보 반환) helm client deployed releaselist 활용
// 	r.router.HandleFunc(nsPrefix+releasePrefix+"/{release-name}", r.Hcm.GetReleases).Methods("GET")
// 	r.router.HandleFunc(nsPrefix+releasePrefix, r.Hcm.InstallRelease).Methods("POST")                       // helm release 생성
// 	r.router.HandleFunc(nsPrefix+releasePrefix+"/{release-name}", r.Hcm.UnInstallRelease).Methods("DELETE") // 설치된 release 전부 삭제 (path-variable : 특정 release 삭제)
// 	r.router.HandleFunc(nsPrefix+releasePrefix+"/{release-name}", r.Hcm.RollbackRelease).Methods("PATCH")   // 일단 미사용 (update / rollback)
// 	http.Handle("/", r.router)
// }

// Get all-namespace 구현
func (hcm *HelmClientManager) GetAllReleases(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get All Releases")
	setResponseHeader(w)

	if err := hcm.SetClientNS(""); err != nil {
		klog.Errorln("Error occurs while setting namespace")
		return
	}

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

	for _, rel := range customReleases {
		response.Release = append(response.Release, *rel)
	}

	respond(w, http.StatusOK, response)
}

func (hcm *HelmClientManager) GetReleases(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("GetRelease")
	setResponseHeader(w)

	vars := mux.Vars(r)
	namespace := vars["ns-name"]

	if err := hcm.SetClientNS(namespace); err != nil {
		klog.Errorln("Error occurs while setting namespace")
		return
	}

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
	setResponseHeader(w)
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

	if err := hcm.SetClientNS(namespace); err != nil {
		klog.Errorln("Error occurs while setting namespace")
		return
	}
	klog.Info(namespace)

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

	if err := hcm.SetClientNS(namespace); err != nil {
		klog.Errorln("Error occurs while setting namespace")
		return
	}

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
	setResponseHeader(w)

	vars := mux.Vars(r)
	namespace := vars["ns-name"]
	reqReleaseName := vars["release-name"]

	if err := hcm.SetClientNS(namespace); err != nil {
		klog.Errorln("Error occurs while setting namespace")
		return
	}

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

func setResponseHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Access-Control-Allow-Headers", "Origin,Accept,X-Requested-With,Content-Type,Access-Control-Request-Method,Access-Control-Request-Headers,Authorization")

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
