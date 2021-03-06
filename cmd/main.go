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

	releasePrefix    = "/releases"
	chartPrefix      = "/charts"
	repoPrefix       = "/repos"
	repositoryPrefix = "/repository"
	nsPrefix         = "/ns/{ns-name}"
	allNamespaces    = "/all-namespaces"
	versionPrefix    = "/versions/{version}"
)

func main() {
	klog.Infoln("initializing server....")

	router := mux.NewRouter()
	apiRouter := router.PathPrefix(helmPrefix).Subrouter()

	hcm := apis.NewHelmClientManager()
	hcm.AddDefaultRepo() // Add default repo

	chartHandler := apis.NewChartHandler(hcm)

	// Repository Test
	apiRouter.HandleFunc(repositoryPrefix, hcm.CreateChartRepo).Methods("POST")

	// Chart API
	apiRouter.HandleFunc(chartPrefix, chartHandler.GetCharts).Methods("GET")                 // 설치 가능한 chart list 반환
	apiRouter.HandleFunc(chartPrefix+"/{chart-name}", chartHandler.GetCharts).Methods("GET") // (query : category 분류된 chart list반환 / path-varaible : 특정 chart data + value.yaml 반환)
	apiRouter.HandleFunc(chartPrefix+"/{chart-name}"+versionPrefix, chartHandler.GetCharts).Methods("GET")

	// // Chart API
	// apiRouter.HandleFunc(chartPrefix, hcm.GetCharts).Methods("GET")                 // 설치 가능한 chart list 반환
	// apiRouter.HandleFunc(chartPrefix+"/{chart-name}", hcm.GetCharts).Methods("GET") // (query : category 분류된 chart list반환 / path-varaible : 특정 chart data + value.yaml 반환)

	// Release API
	apiRouter.HandleFunc(allNamespaces+releasePrefix, hcm.GetAllReleases).Methods("GET") // all-namespaces 릴리즈 반환
	apiRouter.HandleFunc(nsPrefix+releasePrefix, hcm.GetReleases).Methods("GET")         // 설치된 release list 반환 (path-variable : 특정 release 정보 반환) helm client deployed releaselist 활용
	apiRouter.HandleFunc(nsPrefix+releasePrefix+"/{release-name}", hcm.GetReleases).Methods("GET")
	apiRouter.HandleFunc(nsPrefix+releasePrefix, hcm.InstallRelease).Methods("POST")                       // helm release 생성
	apiRouter.HandleFunc(nsPrefix+releasePrefix+"/{release-name}", hcm.UnInstallRelease).Methods("DELETE") // 설치된 release 전부 삭제 (path-variable : 특정 release 삭제)
	apiRouter.HandleFunc(nsPrefix+releasePrefix+"/{release-name}", hcm.RollbackRelease).Methods("PATCH")   // 일단 미사용 (update / rollback)

	// Repo API
	apiRouter.HandleFunc(repoPrefix, chartHandler.GetChartRepos).Methods("GET")                     // 현재 추가된 Helm repo list 반환
	apiRouter.HandleFunc(repoPrefix, chartHandler.AddChartRepo).Methods("POST")                     // Helm repo 추가
	apiRouter.HandleFunc(repoPrefix, chartHandler.UpdateChartRepo).Methods("PUT")                   // Helm repo sync 맞추기
	apiRouter.HandleFunc(repoPrefix+"/{repo-name}", chartHandler.DeleteChartRepo).Methods("DELETE") // repo-name의 Repo 삭제 (index.yaml과 )

	http.Handle("/", router)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", 8081), nil); err != nil {
		klog.Errorln(err, "failed to initialize a server")
	}

}
