package release

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"strings"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/klog"

	"github.com/tmax-cloud/helm-apiserver/internal/utils"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"

	helmclient "github.com/mittwald/go-helm-client"
)

func (sh *ReleaseHandler) releaseHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	_, ns := vars["ns-name"]
	_, rel := vars["release-name"]

	switch {
	case ns && !rel:
		if req.Method == http.MethodPost {
			sh.InstallRelease(w, req)
		}

		if req.Method == http.MethodGet {
			sh.GetReleases(w, req)
		}
	case ns && rel:
		if req.Method == http.MethodPut {
			sh.UpgradeRelease(w, req)
		}
		if req.Method == http.MethodDelete {
			sh.UnInstallRelease(w, req)
		}
		if req.Method == http.MethodGet {
			sh.GetReleases(w, req)
		}
	case !ns:
		sh.GetAllReleases(w, req)
	}
}

// Get all-namespace 구현
func (sh *ReleaseHandler) GetAllReleases(w http.ResponseWriter, r *http.Request) {
	utils.SetResponseHeader(w)

	// bearerToken := r.Header.Get("authorization")

	// All NameSpace 설정
	if err := sh.hcm.SetClientNS(""); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return
	}

	// Release List 받아오기
	releases, err := sh.hcm.Hci.ListDeployedReleases()
	if err != nil {
		klog.V(1).Info(err, "failed to get helm release list")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
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
			klog.V(1).Info(err, "Error occurs while setting Obejcts field")
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

	utils.Respond(w, http.StatusOK, response)
}

func (sh *ReleaseHandler) GetReleases(w http.ResponseWriter, r *http.Request) {
	utils.SetResponseHeader(w)
	// bearerToken := r.Header.Get("authorization")

	vars := mux.Vars(r)
	namespace := vars["ns-name"]

	if err := sh.hcm.SetClientNS(namespace); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return
	}

	releases, err := sh.hcm.Hci.ListDeployedReleases()
	if err != nil {
		klog.V(1).Info(err, "failed to get helm release list")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
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
			klog.V(1).Info(err, "Error occurs while setting Obejcts field")
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

		utils.Respond(w, http.StatusOK, response)
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

	utils.Respond(w, http.StatusOK, response)
}

func (sh *ReleaseHandler) GetReleasesForWS(ns string) *schemas.ReleaseResponse {
	// bearerToken := r.Header.Get("authorization")

	if err := sh.hcm.SetClientNS(ns); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return nil
	}

	releases, err := sh.hcm.Hci.ListDeployedReleases()
	if err != nil {
		klog.V(1).Info(err, "failed to get helm release list")
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
			klog.V(1).Info(err, "Error occurs while setting Obejcts field")
		}
	}

	response := &schemas.ReleaseResponse{}

	for _, rel := range customReleases {
		response.Release = append(response.Release, *rel)
	}

	return response
}

func (sh *ReleaseHandler) InstallRelease(w http.ResponseWriter, r *http.Request) {
	utils.SetResponseHeader(w)
	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.V(1).Info(err, "failed to decode request")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

	vars := mux.Vars(r)
	namespace := vars["ns-name"]
	bearerToken := r.Header.Get("authorization")
	tokenValue := strings.TrimPrefix(bearerToken, "Bearer ")
	if err := sh.hcm.SetClientTokenNS(namespace, tokenValue); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return
	}

	values := trimComments(req.Values)

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.ReleaseName,
		ChartName:   req.PackageURL,
		Namespace:   namespace,
		ValuesYaml:  values,
		Version:     req.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	releases, err := sh.hcm.Hci.ListDeployedReleases()
	if err != nil {
		klog.V(1).Info(err, "failed to get helm release list")
	}
	for _, rel := range releases {
		if rel.Name == req.ReleaseName {
			utils.Respond(w, http.StatusBadRequest, &schemas.Error{
				Description: req.ReleaseName + " release is already exist",
			})
			return
		}
	}

	if _, err := sh.hcm.Hci.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
		klog.V(1).Info(err, "failed to install release")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while installing helm release",
		})
		return
	}

	utils.Respond(w, http.StatusOK, req.ReleaseName+" release is successfully installed")

	if err := sh.hcm.SetDefaultToken(namespace); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return
	}

	// 일단 all-ns 로 보내고 hub에서 filtering
	if len(hub.clients) > 0 {
		releaseList := sh.GetReleasesForWS("")
		hub.broadcast <- *releaseList
	}
	removeCacheFile()
}

func (sh *ReleaseHandler) UpgradeRelease(w http.ResponseWriter, r *http.Request) {
	utils.SetResponseHeader(w)
	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.V(1).Info(err, "failed to decode request")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

	vars := mux.Vars(r)
	namespace := vars["ns-name"]
	bearerToken := r.Header.Get("authorization")
	tokenValue := strings.TrimPrefix(bearerToken, "Bearer ")
	if err := sh.hcm.SetClientTokenNS(namespace, tokenValue); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return
	}

	values := trimComments(req.Values)

	chartSpec := helmclient.ChartSpec{
		ReleaseName: req.ReleaseName,
		ChartName:   req.PackageURL,
		Namespace:   namespace,
		ValuesYaml:  values,
		Version:     req.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	if _, err := sh.hcm.Hci.InstallOrUpgradeChart(context.Background(), &chartSpec); err != nil {
		klog.V(1).Info(err, "failed to upgrade release")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while upgrading helm release",
		})
		return
	}

	utils.Respond(w, http.StatusOK, req.ReleaseName+" release is successfully upgraded")

	if err := sh.hcm.SetDefaultToken(namespace); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return
	}

	// 일단 all-ns 로 보내고 hub에서 filtering
	if len(hub.clients) > 0 {
		releaseList := sh.GetReleasesForWS("")
		hub.broadcast <- *releaseList
	}
	removeCacheFile()
}

// 일단 안씀
func (sh *ReleaseHandler) RollbackRelease(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// bearerToken := r.Header.Get("authorization")
	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.V(1).Info(err, "failed to decode request")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

	vars := mux.Vars(r)
	namespace := vars["ns-name"]
	reqReleaseName := vars["release-name"]

	if err := sh.hcm.SetClientNS(namespace); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: reqReleaseName,
		ChartName:   req.PackageURL,
		Namespace:   namespace,
		ValuesYaml:  string(req.Values),
		Version:     req.Version,
		UpgradeCRDs: true,
		Wait:        false,
	}

	if err := sh.hcm.Hci.RollbackRelease(&chartSpec, 0); err != nil {
		klog.V(1).Info(err, "failed to rollback chart")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while rollback helm release",
		})
	}

	utils.Respond(w, http.StatusOK, reqReleaseName+" release is successfully rollbacked")
}

func (sh *ReleaseHandler) UnInstallRelease(w http.ResponseWriter, r *http.Request) {
	utils.SetResponseHeader(w)

	vars := mux.Vars(r)
	reqReleaseName := vars["release-name"]
	namespace := vars["ns-name"]
	bearerToken := r.Header.Get("authorization")
	tokenValue := strings.TrimPrefix(bearerToken, "Bearer ")
	if err := sh.hcm.SetClientTokenNS(namespace, tokenValue); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return
	}

	if err := sh.hcm.Hci.UninstallReleaseByName(reqReleaseName); err != nil {
		klog.V(1).Info(err, "failed to uninstall chart")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while uninstalling helm release",
		})
		return
	}

	utils.Respond(w, http.StatusOK, reqReleaseName+" release is successfully uninstalled")

	if err := sh.hcm.SetDefaultToken(namespace); err != nil {
		klog.V(1).Info(err, "Error occurs while setting namespace")
		return
	}

	// 일단 all-ns 로 보내고 hub에서 filtering
	if len(hub.clients) > 0 {
		releaseList := sh.GetReleasesForWS("")
		hub.broadcast <- *releaseList
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

func trimComments(input string) string {
	// values 시작 부분 특수문자 제거
	re := regexp.MustCompile(`[|<>]+`)
	key := re.ReplaceAllString(input, "")

	// remove comments in values.yaml file
	temp := make(map[string]interface{})
	_ = yamlv3.Unmarshal([]byte(key), &temp)
	values, _ := yamlv3.Marshal(temp)

	return string(values)
}

func removeCacheFile() {
	fs, _ := os.ReadDir(repositoryCache)
	for _, f := range fs {
		if strings.HasSuffix(f.Name(), "tgz") {
			os.Remove(repositoryCache + "/" + f.Name())
		}
	}
}
