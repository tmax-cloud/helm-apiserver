package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/pflag"
	"github.com/tmax-cloud/helm-apiserver/internal"
	"github.com/tmax-cloud/helm-apiserver/pkg/apis"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	releasePrefix = "/releases"
	chartPrefix   = "/charts"
	repoPrefix    = "/repos"

	repositoryCache  = "/tmp/.helmcache" // 캐시 디렉토리. 특정 chart-repo에 해당하는 chart 이름 리스트 txt파일과, 해당 repo의 index.yaml 파일이 저장됨
	repositoryConfig = "/tmp/.helmrepo"  // 현재 add된 repo들 저장. go helm client 버그. 무조건 /tmp/.helmrepo 에다가 저장됨.
)

func main() {
	klog.Infoln("initializing server....")
	router := mux.NewRouter()

	// Create HelmClientManager
	hcm := apis.HelmClientManager{Hci: newHelmClientInterface(), Hcs: newHelmClientStruct()}

	router.HandleFunc(chartPrefix, hcm.GetCharts).Methods("GET")                 // 설치 가능한 chart list 반환
	router.HandleFunc(chartPrefix+"/{chart-name}", hcm.GetCharts).Methods("GET") // (query : category 분류된 chart list반환 / path-varaible : 특정 chart data + value.yaml 반환)

	router.HandleFunc(releasePrefix, hcm.GetReleases).Methods("GET") // 설치된 release list 반환 (path-variable : 특정 release 정보 반환) helm client deployed releaselist 활용
	router.HandleFunc(releasePrefix+"/{release-name}", hcm.GetReleases).Methods("GET")
	router.HandleFunc(releasePrefix, hcm.InstallRelease).Methods("POST")                       // helm release 생성
	router.HandleFunc(releasePrefix+"/{release-name}", hcm.UnInstallRelease).Methods("DELETE") // 설치된 release 전부 삭제 (path-variable : 특정 release 삭제)
	router.HandleFunc(releasePrefix+"/{release-name}", hcm.RollbackRelease).Methods("PATCH")   // 일단 미사용 (update / rollback)

	router.HandleFunc(repoPrefix, hcm.GetChartRepos).Methods("GET")                     // 현재 추가된 Helm repo list 반환
	router.HandleFunc(repoPrefix, hcm.AddChartRepo).Methods("POST")                     // Helm repo 추가
	router.HandleFunc(repoPrefix, hcm.UpdateChartRepo).Methods("PUT")                   // Helm repo sync 맞추기
	router.HandleFunc(repoPrefix+"/{repo-name}", hcm.DeleteChartRepo).Methods("DELETE") // repo-name의 Repo 삭제 (index.yaml과 )

	http.Handle("/", router)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", 8081), nil); err != nil {
		klog.Errorln(err, "failed to initialize a server")
	}

}

func newHelmClientInterface() helmclient.Client {
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Errorln(err, "failed to get rest config")
	}

	c, _ := client.New(cfg, client.Options{})

	sa, _ := internal.GetServiceAccount(c, types.NamespacedName{Name: "helm-server-test-sa", Namespace: "helm-ns"})
	var secretName string

	for _, sec := range sa.Secrets {
		secretName = sec.Name
	}

	testSecret, _ := internal.GetSecret(c, types.NamespacedName{Name: secretName, Namespace: "helm-ns"})
	token := testSecret.Data["token"]

	opt := &helmclient.Options{
		RepositoryCache:  repositoryCache,
		RepositoryConfig: repositoryConfig,
		Debug:            true,
		Linting:          true,
	}

	cfg.BearerToken = string(token)
	cfg.BearerTokenFile = ""

	helmClientInterface, err := helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{Options: opt, RestConfig: cfg})
	if err != nil {
		klog.Errorln(err, "failed to create helm client")
	}

	return helmClientInterface
}

func newHelmClientStruct() helmclient.HelmClient {
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Errorln(err, "failed to get rest config")
	}

	opt := &helmclient.Options{
		RepositoryCache:  repositoryCache,
		RepositoryConfig: repositoryConfig,
		Debug:            true,
		Linting:          true,
	}

	newHelmClientStruct, err := newClientFromRestConf(&helmclient.RestConfClientOptions{Options: opt, RestConfig: cfg})
	if err != nil {
		klog.Errorln(err, "failed to create helm client")
	}

	return newHelmClientStruct
}

// NewClientFromRestConf returns a new Helm client constructed with the provided REST config options.
func newClientFromRestConf(options *helmclient.RestConfClientOptions) (helmclient.HelmClient, error) {
	settings := cli.New()

	clientGetter := newRESTClientGetter(options.Namespace, nil, options.RestConfig)

	return newClient(options.Options, clientGetter, settings)
}

// field 정의 안되는 문제 확인 필요
func newRESTClientGetter(namespace string, kubeConfig []byte, restConfig *rest.Config) *helmclient.RESTClientGetter {
	return &helmclient.RESTClientGetter{}
}

func newClient(options *helmclient.Options, clientGetter genericclioptions.RESTClientGetter, settings *cli.EnvSettings) (helmclient.HelmClient, error) {
	setEnvSettings(options, settings)

	debugLog := options.DebugLog

	actionConfig := new(action.Configuration)
	actionConfig.Init(
		clientGetter,
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		debugLog,
	)

	return helmclient.HelmClient{
		Settings:     settings,
		Providers:    getter.All(settings),
		ActionConfig: actionConfig,
		DebugLog:     debugLog,
	}, nil
}

func setEnvSettings(options *helmclient.Options, settings *cli.EnvSettings) error {

	// set the namespace with this ugly workaround because cli.EnvSettings.namespace is private
	// thank you helm!
	if options.Namespace != "" {
		pflags := pflag.NewFlagSet("", pflag.ContinueOnError)
		settings.AddFlags(pflags)
		err := pflags.Parse([]string{"-n", options.Namespace})
		if err != nil {
			return err
		}
	}

	settings.RepositoryCache = options.RepositoryCache
	settings.RepositoryConfig = options.RepositoryConfig
	settings.Debug = options.Debug

	return nil
}
