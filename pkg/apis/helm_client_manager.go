package apis

import (
	"os"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/tmax-cloud/helm-apiserver/internal"

	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	repositoryCache  = "/tmp/.helmcache" // 캐시 디렉토리. 특정 chart-repo에 해당하는 chart 이름 리스트 txt파일과, 해당 repo의 index.yaml 파일이 저장됨
	repositoryConfig = "/tmp/.helmrepo"  // 현재 add된 repo들 저장. go helm client 버그. 무조건 /tmp/.helmrepo 에다가 저장됨.
)

type HelmClientManager struct {
	Hci helmclient.Client
	Hcs helmclient.HelmClient
}

func (hcm *HelmClientManager) Init() {
	hcm.Hci = internal.NewHelmClientInterface()
	hcm.Hcs = helmclient.HelmClient{}
}

func (hcm *HelmClientManager) Init2() {
	hcm.Hcs = *typeAssert(hcm.Hci)
	hcm.Hci = &hcm.Hcs
}

func typeAssert(i helmclient.Client) *helmclient.HelmClient {
	ret, ok := i.(*helmclient.HelmClient)
	if ok {
		return ret
	}
	return nil
}

func (hcm *HelmClientManager) SetClientNS(ns string) error {

	cfg, err := hcm.Hcs.ActionConfig.RESTClientGetter.ToRESTConfig()
	if err != nil {
		klog.Errorln(err, "failed to get rest config")
	}

	settings := hcm.Hcs.Settings
	options := &helmclient.Options{
		Namespace:        ns,
		RepositoryCache:  repositoryCache,
		RepositoryConfig: repositoryConfig,
		Debug:            true,
		Linting:          true,
	}

	clientGetter := helmclient.NewRESTClientGetter(ns, nil, cfg)

	if err = internal.SetEnvSettings(options, settings); err != nil {
		klog.Errorln(err, "failed to set Env settings")
		return err
	}

	if err = hcm.Hcs.ActionConfig.Init(
		clientGetter,
		ns,
		os.Getenv("HELM_DRIVER"),
		hcm.Hcs.ActionConfig.Log,
	); err != nil {
		klog.Errorln(err, "failed to init action config")
		return err
	}

	hcm.Hcs.Providers = getter.All(settings)
	return nil
}

func (hcm *HelmClientManager) SetClientTLS(serverName string) {
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

	// cfg.BearerToken = token
	// cfg.BearerTokenFile = ""
	cfg.TLSClientConfig.CertFile = public_key
	cfg.TLSClientConfig.ServerName = serverName

	helmClient, err := helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{Options: opt, RestConfig: cfg})
	if err != nil {
		klog.Errorln(err, "failed to create helm client")
	}

	hcm.Hci = helmClient
}

// [TODO]: req.header로 받은 token 값으로 교체 예정
// func (hcm *HelmClientManager) SetClientToken(ns string) {}

// cfg, err := config.GetConfig()
// if err != nil {
// 	klog.Errorln(err, "failed to get rest config")
// }

// cfg, err := hcm.Hcs.ActionConfig.RESTClientGetter.ToRESTConfig()
// if err != nil {
// 	klog.Errorln(err, "failed to get rest config")
// }

// opt := &helmclient.Options{
// 	Namespace:        ns,
// 	RepositoryCache:  repositoryCache,
// 	RepositoryConfig: repositoryConfig,
// 	Debug:            true,
// 	Linting:          true,
// }

// cfg.BearerToken = token
// cfg.BearerTokenFile = ""
