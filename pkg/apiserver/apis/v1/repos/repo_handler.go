package repos

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

	"github.com/tmax-cloud/helm-apiserver/internal/utils"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
)

// // [Plan A]
// // Helm cache file or dir 을 비워주고 add chart repo

// // [Plan B] - 현재 구현
// // -index.yaml 과 .helmrepo 파일의 sync가 안맞음
// // add chart repo 후, -index.yaml 파일은 있는데 같은이름이 .helmrepo 파일에 없을경우
// // .helmrepo 파일에 request로 들어온 name / url을 덮어씌워주고 마무리.

func (rh *RepoHandler) repoHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	_, rel := vars["repo-name"]

	switch {
	case rel:
		if req.Method == http.MethodGet {
			rh.GetChartRepos(w, req)
		}
		if req.Method == http.MethodPut {
			rh.UpdateChartRepo(w, req)
			rh.UpdateChartCache()
			rh.updateRepoCache()
		}
		if req.Method == http.MethodDelete {
			rh.DeleteChartRepo(w, req)
			rh.UpdateChartCache()
			rh.updateRepoCache()
		}
	case !rel:
		if req.Method == http.MethodPost {
			rh.AddChartRepo(w, req)
			rh.UpdateChartCache()
			rh.updateRepoCache()
		}
		if req.Method == http.MethodGet {
			rh.GetChartRepos(w, req)
		}
	}
}

func (rh *RepoHandler) UpdateChartCache() {
	rh.Index = utils.GetIndex()
	rh.SingleChartEntries = utils.GetSingleChart(rh.Index)
	klog.V(3).Info("Updating ChartCache is done")
}

func (rh *RepoHandler) addDefaultRepo() {
}

// 	rh.log.Info("Add default Chart repo")

// Read repositoryConfig File which contains repo Info list
// repoList, _ := readRepoList()

// for _, repoInfo := range repoList.Repositories {
// 	if repoInfo.Name == "tmax-stable" {
// 		return
// 	}

// }

// chartRepo := repo.Entry{
// 	Name: "tmax-stable",
// 	URL:  "https://tmax-cloud.github.io/helm-charts/stable",
// }

// if err := rh.hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
// 	klog.Errorln(err, "failed to add default tmax chart repo")
// 	return
// }

// if err := InsertChartsToDB(ch.col, chartRepo.Name, chartRepo.URL); err != nil {
// 	klog.Errorln(err, "inserting charts to DB is failed")
// }

// }

func (rh *RepoHandler) AddChartRepo(w http.ResponseWriter, r *http.Request) {
	utils.SetResponseHeader(w)
	req := &schemas.RepoRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.V(1).Info(err, "failed to decode request")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while decoding request",
		})
		return
	}

	// Read repositoryConfig File which contains repo Info list
	repoList, _ := utils.ReadRepoList()
	if repoList != nil {
		var repoNames []string
		// store repo names into repoNames slice
		for _, repoInfo := range repoList.Repositories {
			repoNames = append(repoNames, repoInfo.Name)
		}

		// Check if req repoName is already exist
		for _, repoName := range repoNames {
			if req.Name == repoName {
				utils.Respond(w, http.StatusBadRequest, &schemas.Error{
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
			Name:                  req.Name,
			URL:                   req.RepoURL,
			Username:              req.Id,
			Password:              req.Password,
			InsecureSkipTLSverify: true,
		}
	} else {
		chartRepo = repo.Entry{
			Name:                  req.Name,
			URL:                   req.RepoURL,
			InsecureSkipTLSverify: true,
		}
	}

	if err := rh.hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
		klog.V(1).Info(err, "failed to add chart repo")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while adding helm repo",
		})
		return
	}

	// Read repositoryConfig File which contains repo Info list
	afterRepoList, err := utils.ReadRepoList()
	if err != nil {
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
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
			utils.Respond(w, http.StatusBadRequest, &schemas.Error{
				Error:       err.Error(),
				Description: "Error occurs while sync repo list file",
			})
			return
		}

	}

	klog.V(3).Info(req.Name + " repo is successfully added")
	utils.Respond(w, http.StatusOK, req.Name+" repo is successfully added")
}

func (rh *RepoHandler) GetChartRepos(w http.ResponseWriter, r *http.Request) {
	utils.SetResponseHeader(w)
	if rh.RepoCache == nil || len(rh.Repositories) == 0 {
		utils.Respond(w, http.StatusOK, &schemas.Error{
			Error:       "No helm repository is added",
			Description: "you need to add at least one helm repository",
		})
		return
	}

	// Set Response with repo Info list
	response := &schemas.RepoResponse{}

	vars := mux.Vars(r)
	reqRepoName, exist := vars["repo-name"]
	if exist {
		for _, repo := range rh.Repositories {
			if repo.Name == reqRepoName {
				response.RepoInfo = append(response.RepoInfo, repo)
			}
		}
		klog.V(3).Info("Get Chart repo is successfully done")
		utils.Respond(w, http.StatusOK, response)
		return
	}

	response.RepoInfo = rh.Repositories
	klog.V(3).Info("Get Chart repo is successfully done")
	utils.Respond(w, http.StatusOK, response)
}

func (rh *RepoHandler) DeleteChartRepo(w http.ResponseWriter, r *http.Request) {
	utils.SetResponseHeader(w)

	vars := mux.Vars(r)
	reqRepoName := vars["repo-name"]

	// Read repositoryConfig File which contains repo Info list
	repoList, err := utils.ReadRepoList()
	if err != nil {
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
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
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while sync repo list file",
		})
		return
	}

	// Remove -charts.txt file
	if err := os.Remove(repositoryCache + "/" + reqRepoName + chartsFileSuffix); err != nil {
		klog.V(1).Info(err, "failed to remove charts.txt file")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while removing charts.txt file",
		})
		return
	}

	// Remove -index.yaml file
	if err := os.Remove(repositoryCache + "/" + reqRepoName + indexFileSuffix); err != nil {
		klog.V(1).Info(err, "failed to remove index.yaml file")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while removing index.yaml file",
		})
		return
	}

	klog.V(3).Info(reqRepoName + " is successfully removed")
	utils.Respond(w, http.StatusOK, reqRepoName+" repo is successfully removed")
}

func (rh *RepoHandler) UpdateChartRepo(w http.ResponseWriter, r *http.Request) {
	utils.SetResponseHeader(w)

	// Read repositoryConfig File which contains repo Info list
	repoList, err := utils.ReadRepoList()
	if err != nil {
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
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

	if err := rh.hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
		klog.V(1).Info(err, "failed to update chart repo")
		utils.Respond(w, http.StatusBadRequest, &schemas.Error{
			Error:       err.Error(),
			Description: "Error occurs while updating helm repo " + chartRepo.Name,
		})
		return
	}
	klog.V(3).Info(chartRepo.Name + " repo is successfully updated")

	// ch.UpdateChartHandler()
	utils.Respond(w, http.StatusOK, "repo update is successfully done")
}

func writeRepoList(repoList *schemas.RepositoryFile) error {
	repoList.Generated = time.Now() // repo 삭제 후 발생하는 버그 방지
	newRepoListFile, err := json.Marshal(repoList)
	if err != nil {
		return err
	}

	newRepoListFileYaml, _ := yaml.JSONToYAML(newRepoListFile) // Should transform Json to Yaml

	// Update repository.yaml file without requested repo
	if err := ioutil.WriteFile(repositoryConfig, newRepoListFileYaml, 0644); err != nil {
		return err
	}

	return nil
}

// func getIndex() *schemas.IndexFile {
// 	// Read repositoryConfig File which contains repo Info list
// 	repoList, err := utils.ReadRepoList()
// 	if err != nil {
// 		klog.Errorln(err, "failed to save index file")
// 		return nil
// 	}

// 	repoInfos := make(map[string]string)
// 	// store repo names
// 	for _, repoInfo := range repoList.Repositories {
// 		repoInfos[repoInfo.Name] = repoInfo.Url
// 	}

// 	index := &schemas.IndexFile{}
// 	allEntries := make(map[string]schemas.ChartVersions)
// 	// col := db.GetMongoDBConnetion() // #######테스트########

// 	// read all index.yaml file and save only Entries
// 	for repoName, repoUrl := range repoInfos {
// 		if index, err = utils.ReadRepoIndex(repoName); err != nil {
// 			klog.Errorln(err, "failed to read index file")
// 		}

// 		// add repo info
// 		for key, charts := range index.Entries {
// 			for _, chart := range charts {
// 				chart.Repo.Name = repoName
// 				chart.Repo.Url = repoUrl
// 				// _, err := db.InsertDoc(col, chart) // #######테스트########
// 				// klog.Info("insert done!")
// 				// if err != nil {
// 				// 	klog.Error(err)
// 				// }
// 			}
// 			allEntries[repoName+"_"+key] = charts // 중복 chart name 가능하도록 repo name과 결합
// 		}
// 	}

// 	// filter := bson.D{{}}
// 	// var test []schemas.ChartVersion
// 	// test, _ = db.FindDoc(col, filter, filter)
// 	// for _, ch := range test {
// 	// 	klog.Info(ch.Name)
// 	// }

// 	index.Entries = allEntries
// 	klog.Info("saving index file is done")
// 	return index
// }

// func getSingleChart(index *schemas.IndexFile) map[string]schemas.ChartVersions {
// 	if index == nil {
// 		return nil
// 	}

// 	singleChartEntries := make(map[string]schemas.ChartVersions)
// 	for key, charts := range index.Entries {
// 		var oneChart []*schemas.ChartVersion
// 		oneChart = append(oneChart, charts[0])
// 		singleChartEntries[key] = oneChart
// 	}
// 	return singleChartEntries
// }

// func (rh *RepoHandler) GetChartRepos(w http.ResponseWriter, r *http.Request) {
// 	utils.SetResponseHeader(w)

// 	// Read repositoryConfig File which contains repo Info list
// 	repoList, _ := utils.ReadRepoList()
// 	if repoList == nil {
// 		utils.Respond(w, http.StatusOK, &schemas.Error{
// 			Error:       "No helm repository is added",
// 			Description: "you need to add at least one helm repository",
// 		})
// 		return
// 	}

// 	// Set Response with repo Info list
// 	totalRepo := &schemas.RepoResponse{}
// 	response := &schemas.RepoResponse{}

// 	// set last updated time
// 	for _, repo := range repoList.Repositories {
// 		r_index, err := utils.ReadRepoIndex(repo.Name)
// 		if err != nil {
// 			klog.V(1).Info(err, "failed to read index file")
// 		}
// 		repo.LastUpdated = r_index.Generated
// 		totalRepo.RepoInfo = append(totalRepo.RepoInfo, repo)
// 	}

// 	vars := mux.Vars(r)
// 	reqRepoName, exist := vars["repo-name"]
// 	if exist {
// 		for _, repo := range totalRepo.RepoInfo {
// 			if repo.Name == reqRepoName {
// 				response.RepoInfo = append(response.RepoInfo, repo)
// 			}
// 		}
// 		klog.V(3).Info("Get Chart repo is successfully done")
// 		utils.Respond(w, http.StatusOK, response)
// 		return
// 	}
// 	response.RepoInfo = totalRepo.RepoInfo

// 	klog.V(3).Info("Get Chart repo is successfully done")
// 	utils.Respond(w, http.StatusOK, response)
// }
