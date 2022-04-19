package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/helm-apiserver/pkg/apis"
	"k8s.io/klog"
)

const (
	helmPrefix = "/helm"

	releasePrefix = "/releases"
	chartPrefix   = "/charts"
	repoPrefix    = "/repos"
	nsPrefix      = "/ns/{ns-name}"
	allNamespaces = "/all-namespaces"
)

func main() {
	klog.Infoln("initializing server....")

	router := mux.NewRouter()
	apiRouter := router.PathPrefix(helmPrefix).Subrouter()

	hcm := &apis.HelmClientManager{}
	hcm.Init()
	hcm.Init2()          // for type assertion
	hcm.AddDefaultRepo() // Add default repo

	apiRouter.HandleFunc(chartPrefix, hcm.GetCharts).Methods("GET")                 // 설치 가능한 chart list 반환
	apiRouter.HandleFunc(chartPrefix+"/{chart-name}", hcm.GetCharts).Methods("GET") // (query : category 분류된 chart list반환 / path-varaible : 특정 chart data + value.yaml 반환)

	apiRouter.HandleFunc(allNamespaces+releasePrefix, hcm.GetAllReleases).Methods("GET") // all-namespaces 릴리즈 반환
	apiRouter.HandleFunc(nsPrefix+releasePrefix, hcm.GetReleases).Methods("GET")         // 설치된 release list 반환 (path-variable : 특정 release 정보 반환) helm client deployed releaselist 활용
	apiRouter.HandleFunc(nsPrefix+releasePrefix+"/{release-name}", hcm.GetReleases).Methods("GET")
	apiRouter.HandleFunc(nsPrefix+releasePrefix, hcm.InstallRelease).Methods("POST")                       // helm release 생성
	apiRouter.HandleFunc(nsPrefix+releasePrefix+"/{release-name}", hcm.UnInstallRelease).Methods("DELETE") // 설치된 release 전부 삭제 (path-variable : 특정 release 삭제)
	apiRouter.HandleFunc(nsPrefix+releasePrefix+"/{release-name}", hcm.RollbackRelease).Methods("PATCH")   // 일단 미사용 (update / rollback)

	apiRouter.HandleFunc(repoPrefix, hcm.GetChartRepos).Methods("GET")                     // 현재 추가된 Helm repo list 반환
	apiRouter.HandleFunc(repoPrefix, hcm.AddChartRepo).Methods("POST")                     // Helm repo 추가
	apiRouter.HandleFunc(repoPrefix, hcm.UpdateChartRepo).Methods("PUT")                   // Helm repo sync 맞추기
	apiRouter.HandleFunc(repoPrefix+"/{repo-name}", hcm.DeleteChartRepo).Methods("DELETE") // repo-name의 Repo 삭제 (index.yaml과 )

	http.Handle("/", router)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", 8081), nil); err != nil {
		klog.Errorln(err, "failed to initialize a server")
	}

}
