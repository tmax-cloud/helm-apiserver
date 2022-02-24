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

// [Plan A]
// Helm cache file or dir 을 비워주고 add chart repo

// [Plan B] - 현재 구현
// -index.yaml 과 .helmrepo 파일의 sync가 안맞음
// add chart repo 후, -index.yaml 파일은 있는데 같은이름이 .helmrepo 파일에 없을경우
// .helmrepo 파일에 request로 들어온 name / url을 덮어씌워주고 마무리.

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

	// Read repositoryConfig File which contains repo Info list
	repoList, err := readRepoList()
	if err != nil {
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
		})
		return
	}

	var repoNames []string
	// store repo names into repoNames slice
	for _, repoInfo := range repoList.Repositories {
		repoNames = append(repoNames, repoInfo.Name)
	}

	// Check if req repoName is already exist
	for _, repoName := range repoNames {
		if req.Name == repoName {
			klog.Errorln(req.Name + " name repository is already exist")
			respond(w, http.StatusBadRequest, &schemas.Error{
				Description: req.Name + " name repository is already exist",
			})
			return
		}
	}

	// TODO : Private Repository도 지원해줘야 함
	chartRepo := repo.Entry{
		Name: req.Name,
		URL:  req.RepoURL,
	}

	if err := hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
		klog.Errorln(err, "failed to add chart repo")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while adding helm repo",
		})
		return
	}

	// Read repositoryConfig File which contains repo Info list
	afterRepoList, err := readRepoList()
	if err != nil {
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
		})
		return
	}

	sync := true
	for _, repo := range afterRepoList.Repositories {
		if repo.Name == req.Name {
			sync = false
		}
	}

	// -index.yaml 파일과 .helmrepo 파일 sync
	if sync {
		afterRepoList.Repositories = append(afterRepoList.Repositories, schemas.Repository{
			Name: req.Name,
			Url:  req.RepoURL,
		})

		if err := writeRepoList(afterRepoList); err != nil {
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while sync repo list file",
			})
			return
		}

	}

	klog.Infoln(req.Name + " repo is successfully added")
	respond(w, http.StatusOK, req.Name+" repo is successfully added")
}

func (hcm *HelmClientManager) GetChartRepos(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get chartRepos")
	w.Header().Set("Content-Type", "application/json")

	// Read repositoryConfig File which contains repo Info list
	repoList, err := readRepoList()
	if err != nil {
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
		})
		return
	}

	// Set Response with repo Info list
	response := &schemas.RepoResponse{}
	response.RepoInfo = repoList.Repositories

	klog.Infoln("Get Chart repo is successfully done")
	respond(w, http.StatusOK, response)
}

func (hcm *HelmClientManager) DeleteChartRepo(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Delete chartRepos")
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	reqRepoName := vars["repo-name"]

	// Read repositoryConfig File which contains repo Info list
	repoList, err := readRepoList()
	if err != nil {
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
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

	if err := writeRepoList(newRepoList); err != nil {
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while sync repo list file",
		})
		return
	}

	// Remove -charts.txt file
	if err := os.Remove(repositoryCache + "/" + reqRepoName + chartsFileSuffix); err != nil {
		klog.Errorln(err, "failed to remove charts.txt file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while removing charts.txt file",
		})
		return
	}

	// Remove -index.yaml file
	if err := os.Remove(repositoryCache + "/" + reqRepoName + indexFileSuffix); err != nil {
		klog.Errorln(err, "failed to remove index.yaml file")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while removing index.yaml file",
		})
		return
	}

	klog.Infoln(reqRepoName + " is successfully removed")
	respond(w, http.StatusOK, reqRepoName+" repo is successfully removed")
}

func (hcm *HelmClientManager) UpdateChartRepo(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Update ChartRepo")
	w.Header().Set("Content-Type", "application/json")

	// Read repositoryConfig File which contains repo Info list
	repoList, err := readRepoList()
	if err != nil {
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
		})
		return
	}

	chartRepo := repo.Entry{}
	for _, repo := range repoList.Repositories {
		chartRepo.Name = repo.Name
		chartRepo.URL = repo.Url

		// TODO : Private Repository도 지원해줘야 함
		if err := hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
			klog.Errorln(err, "failed to update chart repo")
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while updating helm repo " + chartRepo.Name,
			})
			return
		}
		klog.Infoln(chartRepo.Name + " repo is successfully updated")
	}

	respond(w, http.StatusOK, "repo update is successfully done")
}

func writeRepoList(repoList *schemas.RepositoryFile) error {
	newRepoListFile, err := json.Marshal(repoList)
	if err != nil {
		klog.Errorln(err, "failed to marshal repo list")
		return err
	}

	newRepoListFileYaml, _ := yaml.JSONToYAML(newRepoListFile) // Should transform Json to Yaml

	// Update repository.yaml file without requested repo
	if err := ioutil.WriteFile(repositoryConfig, newRepoListFileYaml, 0644); err != nil {
		klog.Errorln(err, "failed to write new repo list file")
		return err
	}

	return nil
}
