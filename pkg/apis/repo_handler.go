package apis

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/time"
	"k8s.io/klog"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tmax-cloud/helm-apiserver/pkg/db"

	// "github.com/tmax-cloud/helm-apiserver/internal"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
)

// [Plan A]
// Helm cache file or dir 을 비워주고 add chart repo

// [Plan B] - 현재 구현
// -index.yaml 과 .helmrepo 파일의 sync가 안맞음
// add chart repo 후, -index.yaml 파일은 있는데 같은이름이 .helmrepo 파일에 없을경우
// .helmrepo 파일에 request로 들어온 name / url을 덮어씌워주고 마무리.

// [TODO]
// GET Repo 에서 {repo-name} 으로 상세 정보 return 필요
// Get repo에서 시간 추가해줄것
// Add Repo에서 name / url check logic 추가
// update API 는 특정 repo만 update 되도록 변경

func (hcm *HelmClientManager) AddDefaultRepo() {
	klog.Infoln("Add default Chart repo")

	// Read repositoryConfig File which contains repo Info list
	// repoList, _ := readRepoList()

	// for _, repoInfo := range repoList.Repositories {
	// 	if repoInfo.Name == "tmax-stable" {
	// 		return
	// 	}
	// }

	chartRepo := repo.Entry{
		Name: "tmax-stable",
		URL:  "https://tmax-cloud.github.io/helm-charts/stable",
	}

	if err := hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
		klog.Errorln(err, "failed to add default tmax chart repo")
		return
	}

	// if err := InsertChartsToDB(ch.col, chartRepo.Name, chartRepo.URL); err != nil {
	// 	klog.Errorln(err, "inserting charts to DB is failed")
	// }

}

func (ch *ChartHandler) AddChartRepo(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Add ChartRepo")
	setResponseHeader(w)
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
	repoList, _ := readRepoList()
	if repoList != nil {
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
	}

	// for new version
	chartRepo := repo.Entry{}
	if req.Is_private { // Default false
		// 시크릿 가져와서 user ID / access TOKEN 추가
		chartRepo = repo.Entry{
			Name:     req.Name,
			URL:      req.RepoURL,
			Username: req.Id,
			Password: req.Password,
			// CAFile:                ca_crt,
			// CertFile:              public_key,
			// KeyFile:               private_key,
			InsecureSkipTLSverify: true,
		}
	} else {
		chartRepo = repo.Entry{
			Name:                  req.Name,
			URL:                   req.RepoURL,
			InsecureSkipTLSverify: true,
		}
	}

	// chartRepo := repo.Entry{
	// 	Name: req.Name,
	// 	URL:  req.RepoURL,
	// 	// CAFile:                ca_crt,
	// 	// CertFile:              public_key,
	// 	// KeyFile:               private_key,
	// 	InsecureSkipTLSverify: true,
	// }

	// hcm.SetClientTLS(req.RepoURL)

	if err := ch.hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
		klog.Errorln(err, "failed to add chart repo")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while adding helm repo",
		})
		return
	}

	// if err := InsertChartsToDB(ch.col, req.Name, req.RepoURL); err != nil {
	// 	klog.Errorln(err, "inserting charts to DB is failed")
	// }

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
	// -index.yaml 파일은 생기는데 .helmrepo 파일 update 안되는 버그 있음
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

	ch.UpdateChartHandler() // 확인 필요

	klog.Infoln(req.Name + " repo is successfully added")
	respond(w, http.StatusOK, req.Name+" repo is successfully added")
}

func (ch *ChartHandler) GetChartRepos(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Get chartRepos")
	setResponseHeader(w)

	// Read repositoryConfig File which contains repo Info list
	repoList, _ := readRepoList()
	if repoList == nil {
		respond(w, http.StatusOK, &schemas.Error{
			Error:       "No helm repository is added",
			Description: "you need to add at least one helm repository",
		})
		return
	}

	// Set Response with repo Info list
	totalRepo := &schemas.RepoResponse{}
	response := &schemas.RepoResponse{}

	// set last updated time
	for _, repo := range repoList.Repositories {
		r_index, err := readRepoIndex(repo.Name)
		if err != nil {
			klog.Errorln(err, "failed to read index file")
		}
		repo.LastUpdated = r_index.Generated
		totalRepo.RepoInfo = append(totalRepo.RepoInfo, repo)
	}

	vars := mux.Vars(r)
	reqRepoName, exist := vars["repo-name"]
	if exist {
		for _, repo := range totalRepo.RepoInfo {
			if repo.Name == reqRepoName {
				response.RepoInfo = append(response.RepoInfo, repo)
			}
		}
		klog.Infoln("Get Chart repo is successfully done")
		respond(w, http.StatusOK, response)
		return
	}
	response.RepoInfo = totalRepo.RepoInfo

	klog.Infoln("Get Chart repo is successfully done")
	respond(w, http.StatusOK, response)
}

func (ch *ChartHandler) DeleteChartRepo(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Delete chartRepos")
	setResponseHeader(w)

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

	ch.UpdateChartHandler()

	klog.Infoln(reqRepoName + " is successfully removed")
	respond(w, http.StatusOK, reqRepoName+" repo is successfully removed")
}

func (ch *ChartHandler) UpdateChartRepo(w http.ResponseWriter, r *http.Request) {
	klog.Infoln("Update ChartRepo")
	setResponseHeader(w)

	// Read repositoryConfig File which contains repo Info list
	repoList, err := readRepoList()
	if err != nil {
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while reading repository list file",
		})
		return
	}

	vars := mux.Vars(r)
	reqRepoName := vars["repo-name"]

	chartRepo := repo.Entry{}
	for _, repo := range repoList.Repositories {
		if repo.Name == reqRepoName {
			chartRepo.Name = repo.Name
			chartRepo.URL = repo.Url
			chartRepo.Username = repo.UserName
			chartRepo.Password = repo.Password
		}
	}

	if err := ch.hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
		klog.Errorln(err, "failed to update chart repo")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while updating helm repo " + chartRepo.Name,
		})
		return
	}
	klog.Infoln(chartRepo.Name + " repo is successfully updated")

	ch.UpdateChartHandler()
	respond(w, http.StatusOK, "repo update is successfully done")
}

func writeRepoList(repoList *schemas.RepositoryFile) error {
	repoList.Generated = time.Now() // repo 삭제 후 발생하는 버그 방지
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

func InsertChartsToDB(col *mongo.Collection, repoName string, repoUrl string) error {

	index, err := readRepoIndex(repoName)
	if err != nil {
		klog.Errorln(err, "failed to read index file")
		return err
	}

	// Insert To DB
	for _, charts := range index.Entries {
		for _, chart := range charts {
			chart.Repo.Name = repoName
			chart.Repo.Url = repoUrl
			_, err := db.InsertDoc(col, chart) // #######테스트########
			klog.Info("insert done!")
			if err != nil {
				klog.Errorln(err)
				return err
			}
		}
	}

	return nil
}
