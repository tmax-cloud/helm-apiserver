package repos

import (
	"fmt"
	"net/http"

	"github.com/tmax-cloud/helm-apiserver/internal/apiserver"
	"github.com/tmax-cloud/helm-apiserver/internal/hclient"
	"github.com/tmax-cloud/helm-apiserver/internal/utils"
	"github.com/tmax-cloud/helm-apiserver/internal/wrapper"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/tmax-cloud/helm-apiserver/pkg/apiserver/apis/v1/chart"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

const (
	indexFileSuffix  = "-index.yaml"
	chartsFileSuffix = "-charts.txt"
	APIVersion       = "v1"

	repositoryCache  = "/tmp/.helmcache" // 캐시 디렉토리. 특정 chart-repo에 해당하는 chart 이름 리스트 txt파일과, 해당 repo의 index.yaml 파일이 저장됨
	repositoryConfig = "/tmp/.helmrepo"  // 현재 add된 repo들 저장.
)

type RepoHandler struct {
	hcm *hclient.HelmClientManager
	*chart.ChartCache
	authorizer apiserver.Authorizer
	*RepoCache
}

type RepoCache struct {
	Repositories []schemas.Repository
}

func NewHandler(parent wrapper.RouterWrapper, hcm *hclient.HelmClientManager, authCli authorization.AuthorizationV1Interface, chartCache *chart.ChartCache) (apiserver.APIHandler, error) {
	repoHandler := &RepoHandler{hcm: hcm, ChartCache: chartCache, RepoCache: nil}
	repoHandler.updateRepoCache()
	repoHandler.initRepoCache()

	// Authorizer
	repoHandler.authorizer = apiserver.NewAuthorizer(authCli, apiserver.APIGroup, APIVersion, "update")

	// /v1/repos
	repoWrapper := wrapper.New(fmt.Sprintf("/%s", "repos"), []string{http.MethodGet, http.MethodPost}, repoHandler.repoHandler)
	if err := parent.Add(repoWrapper); err != nil {
		return nil, err
	}
	// /v1/repos/<repo-name>
	repoParamWrapper := wrapper.New(fmt.Sprintf("/%s/{%s}", "repos", "repo-name"), []string{http.MethodGet, http.MethodPut, http.MethodDelete}, repoHandler.repoHandler)
	if err := parent.Add(repoParamWrapper); err != nil {
		return nil, err
	}
	// repoWrapper.Router().Use(repoHandler.authorizer.Authorize)

	return repoHandler, nil
}

func (rh *RepoHandler) updateRepoCache() {
	var repoInfos []schemas.Repository
	repoList, _ := utils.ReadRepoList()
	if repoList == nil {
		return
	}

	for _, repo := range repoList.Repositories {
		r_index, _ := utils.ReadRepoIndex(repo.Name)
		if r_index == nil {
			continue
		}
		repo.LastUpdated = r_index.Generated
		repoInfos = append(repoInfos, repo)
	}

	repoCache := &RepoCache{Repositories: repoInfos}
	rh.RepoCache = repoCache
}

// 재기동시 레포 사라지는 버그 방지
func (rh *RepoHandler) initRepoCache() {
	chartRepo := repo.Entry{}
	for _, re := range rh.Repositories {
		chartRepo.Name = re.Name
		chartRepo.URL = re.Url
		chartRepo.Username = re.UserName
		chartRepo.Password = re.Password
		chartRepo.InsecureSkipTLSverify = true
		rh.hcm.Hci.AddOrUpdateChartRepo(chartRepo)
	}
}
