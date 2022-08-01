package apis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"strings"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"

	"k8s.io/klog"

	helmclient "github.com/mittwald/go-helm-client"
)

// Get all-namespace 구현
func (hcm *HelmClientManager) GetAllReleases(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get All Releases")
	setResponseHeader(w)

	// bearerToken := r.Header.Get("authorization")

	// All NameSpace 설정
	if err := hcm.SetClientNS(""); err != nil {
		klog.Errorln("Error occurs while setting namespace")
		return
	}

	// Release List 받아오기
	releases, err := hcm.Hci.ListDeployedReleases()
	if err != nil {
		klog.Errorln(err, "failed to get helm release list")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while getting helm release list",
		})
		return
	}

	// Releases의 Objects 필드 추가를 위해 customReleases 사용
	var customReleases []*schemas.Release
	for _, rel := range releases {
		customRelease := &schemas.Release{}
		copier.Copy(customRelease, rel)
		customReleases = append(customReleases, customRelease)
	}

	// customReleases에 Objects 필드 값 추가
	for _, rel := range customReleases {
		if err := setObjectsInfo(rel); err != nil {
			klog.Errorln(err, "Error occurs while setting Obejcts field")
		}
	}

	// search releases
	var responseReleases []*schemas.Release
	query, _ := url.ParseQuery(r.URL.RawQuery)
	_, search := query["search"]
	if search {
		searcher := query.Get("search")
		responseReleases = searchRelease(searcher, customReleases)
	} else {
		responseReleases = customReleases
	}

	response := &schemas.ReleaseResponse{}
	for _, rel := range responseReleases {
		response.Release = append(response.Release, *rel)
	}

	respond(w, http.StatusOK, response)
}

func (hcm *HelmClientManager) GetReleases(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("GetRelease")
	setResponseHeader(w)
	// bearerToken := r.Header.Get("authorization")
	// klog.Info(tokenTest)

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

	// search releases
	var responseReleases []*schemas.Release
	query, _ := url.ParseQuery(r.URL.RawQuery)
	_, search := query["search"]
	if search {
		searcher := query.Get("search")
		responseReleases = searchRelease(searcher, customReleases)
	} else {
		responseReleases = customReleases
	}

	for _, rel := range responseReleases {
		response.Release = append(response.Release, *rel)
	}

	respond(w, http.StatusOK, response)
}

func (hcm *HelmClientManager) GetReleasesForWS(ns string) *schemas.ReleaseResponse {
	klog.Infoln("GetRelease")
	// bearerToken := r.Header.Get("authorization")
	// klog.Info(tokenTest)

	if err := hcm.SetClientNS(ns); err != nil {
		klog.Errorln("Error occurs while setting namespace")
		return nil
	}

	releases, err := hcm.Hci.ListDeployedReleases()
	if err != nil {
		klog.Errorln(err, "failed to get helm release list")
		return nil
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

	return response
}

func (hcm *HelmClientManager) InstallRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("InstallRelease")
	setResponseHeader(w)
	// bearerToken := r.Header.Get("authorization")
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

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.ReleaseName,
		ChartName:   req.PackageURL,
		Namespace:   namespace,
		ValuesYaml:  req.Values,
		Version:     req.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	releases, err := hcm.Hci.ListDeployedReleases()
	if err != nil {
		klog.Errorln(err, "failed to get helm release list")
	}
	for _, rel := range releases {
		if rel.Name == req.ReleaseName {
			respond(w, http.StatusBadRequest, &schemas.Error{
				Description: req.ReleaseName + " release is already exist",
			})
			return
		}
	}

	if _, err := hcm.Hci.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
		klog.Errorln(err, "failed to install release")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while installing helm release",
		})
		return
	}

	respond(w, http.StatusOK, req.ReleaseName+" release is successfully installed")

	// 일단 all-ns 로 보내고 hub에서 filtering
	if len(hub.clients) > 0 {
		releaseList := hcm.GetReleasesForWS("")
		hub.broadcast <- *releaseList
	}

}

func (hcm *HelmClientManager) UpgradeRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Upgrade Release")
	setResponseHeader(w)
	// bearerToken := r.Header.Get("authorization")
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

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.ReleaseName,
		ChartName:   req.PackageURL,
		Namespace:   namespace,
		ValuesYaml:  req.Values,
		Version:     req.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	// [TODO] Release 중복 이름 check
	if _, err := hcm.Hci.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
		klog.Errorln(err, "failed to upgrade release")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while upgrading helm release",
		})
		return
	}

	respond(w, http.StatusOK, req.ReleaseName+" release is successfully upgraded")

	// 일단 all-ns 로 보내고 hub에서 filtering
	if len(hub.clients) > 0 {
		releaseList := hcm.GetReleasesForWS("")
		hub.broadcast <- *releaseList
	}

}

// 일단 안씀
func (hcm *HelmClientManager) RollbackRelease(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Rollback Release")
	w.Header().Set("Content-Type", "application/json")
	// bearerToken := r.Header.Get("authorization")
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
	// bearerToken := r.Header.Get("authorization")

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

	// 일단 all-ns 로 보내고 hub에서 filtering
	if len(hub.clients) > 0 {
		releaseList := hcm.GetReleasesForWS("")
		hub.broadcast <- *releaseList

	}
}

func respond(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		klog.Errorln(err, "Error occurs while encoding response body")
	}
}

// CORS 에러로 인해 header 추가 21.04.19
func setResponseHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS") // Method 각각 써줘야 할것 같음
	w.Header().Set("Access-Control-Allow-Headers", "X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version")

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

	rel.Objects = objMap
	return nil
}

// Releases의 이름 or 참조한 차트 이름으로 검색
func searchRelease(searcher string, input []*schemas.Release) (output []*schemas.Release) {
	var filtered []*schemas.Release
	for _, i := range input {
		if strings.Contains(i.Name, searcher) || strings.Contains(i.Chart.Name(), searcher) {
			filtered = append(filtered, i)
		}
	}
	return filtered
}
