package apis

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/klog"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
)

// const (
// 	repositoryCache  = "/tmp/.helmcache" // 캐시 디렉토리. 특정 chart-repo에 해당하는 chart 이름 리스트 txt파일과, 해당 repo의 index.yaml 파일이 저장됨
// 	repositoryConfig = "/tmp/.helmrepo"  // 현재 add된 repo들 저장. go helm client 버그. 무조건 /tmp/.helmrepo 에다가 저장됨.

//  indexFileSuffix  = "-index.yaml"
// 	chartsFileSuffix = "-charts.txt"
// )

func (hcm *HelmClientManager) AddChartRepo(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Add ChartRepo")
	w.Header().Set("Content-Type", "application/json")
	req := &schemas.RepoRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

	// TODO : Private Repository도 지원해줘야 함
	chartRepo := repo.Entry{
		Name: req.Name,
		URL:  req.RepoURL,
	}

	if err := hcm.Hc.AddOrUpdateChartRepo(chartRepo); err != nil {
		klog.Errorln(err, "failed to add chart repo")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while adding helm repo",
		})
		return
	}

	respond(w, http.StatusOK, req.Name+" repo is successfully added")
}

func (hcm *HelmClientManager) GetChartRepos(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get chartRepos")
	w.Header().Set("Content-Type", "application/json")
	req := &schemas.RepoRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

	// Read repositoryConfig File which contains repo Info list
	repoList := &schemas.RepositoryFile{}
	repoListFile, err := ioutil.ReadFile(repositoryConfig)
	if err != nil {
		klog.Errorln(err, "failed to get repository list file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
		})
		return
	}

	repoListFileJson, _ := yaml.YAMLToJSON(repoListFile) // Should transform yaml to Json

	if err := json.Unmarshal(repoListFileJson, repoList); err != nil {
		klog.Errorln(err, "failed to unmarshal repo file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while unmarshalling request",
		})
		return
	}

	// Set Response with repo Info list
	response := &schemas.RepoResponse{}
	for _, repo := range repoList.Repositories { // check 필요
		response.RepoInfo = append(response.RepoInfo, repo)
	}

	respond(w, http.StatusOK, response)
}

func (hcm *HelmClientManager) DeleteChartRepo(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Delete chartRepos")
	w.Header().Set("Content-Type", "application/json")
	req := &schemas.RepoRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

	vars := mux.Vars(r)
	reqRepoName := vars["repo-name"]

	// Read repositoryConfig File which contains repo Info list
	repoList := &schemas.RepositoryFile{}
	repoListFile, err := ioutil.ReadFile(repositoryConfig)
	if err != nil {
		klog.Errorln(err, "failed to get repository list file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
		})
		return
	}

	repoListFileJson, _ := yaml.YAMLToJSON(repoListFile) // Should transform yaml to Json

	if err := json.Unmarshal(repoListFileJson, repoList); err != nil {
		klog.Errorln(err, "failed to unmarshal repo list")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while unmarshalling repo list",
		})
		return
	}

	// Replace repo list without requested repo
	newRepoList := &schemas.RepositoryFile{}
	for _, repo := range repoList.Repositories {
		if repo.Name != reqRepoName {
			newRepoList.Repositories = append(newRepoList.Repositories, repo)
		}
	}

	newRepoListFile, err := json.Marshal(newRepoList)
	if err != nil {
		klog.Errorln(err, "failed to marshal repo list")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while marshalling repo list",
		})
		return
	}

	// Update repository.yaml file without requested repo
	if err := ioutil.WriteFile(repositoryConfig, newRepoListFile, 0644); err != nil {
		klog.Errorln(err, "failed to write new repo list file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while writing new repo list file",
		})
		return
	}

	// [TODO] : File open 먼저 해야되는지 check 필요
	// Remove charts.txt file
	if err := os.Remove(repositoryCache + "/" + reqRepoName + chartsFileSuffix); err != nil {
		klog.Errorln(err, "failed to remove charts.txt file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while removing charts.txt file",
		})
		return
	}

	// Remove index.yaml file
	if err := os.Remove(repositoryCache + "/" + reqRepoName + indexFileSuffix); err != nil {
		klog.Errorln(err, "failed to remove index.yaml file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while removing index.yaml file",
		})
		return
	}

	respond(w, http.StatusOK, reqRepoName+" is successfully removed")
}
