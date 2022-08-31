package hclient

import (
	"os"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/tmax-cloud/helm-apiserver/internal"
	"github.com/tmax-cloud/helm-apiserver/internal/utils"

	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/klog"
)

const (
	repositoryCache  = "/tmp/.helmcache"
	repositoryConfig = "/tmp/.helmrepo"
)

type HelmClientManager struct {
	Hci helmclient.Client
	Hcs helmclient.HelmClient
}

func NewHelmClientManager() *HelmClientManager {
	hci := internal.NewHelmClientInterface()
	hcs := *typeAssert(hci)
	return &HelmClientManager{
		Hci: &hcs,
		Hcs: hcs,
	}
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

func (hcm *HelmClientManager) SetClientTokenNS(ns string, token string) error {

	cfg, err := hcm.Hcs.ActionConfig.RESTClientGetter.ToRESTConfig()
	if err != nil {
		klog.Errorln(err, "failed to get rest config")
	}

	// [TODO] 나머지 테스트 완료 후 활성화
	cfg.BearerToken = token
	cfg.BearerTokenFile = ""

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
		klog.V(1).Info(err, "failed to set Env settings")
		return err
	}

	if err = hcm.Hcs.ActionConfig.Init(
		clientGetter,
		ns,
		os.Getenv("HELM_DRIVER"),
		hcm.Hcs.ActionConfig.Log,
	); err != nil {
		klog.V(1).Info(err, "failed to init action config")
		return err
	}

	hcm.Hcs.Providers = getter.All(settings)
	return nil
}

func (hcm *HelmClientManager) SetDefaultToken(ns string) error {

	cfg, err := hcm.Hcs.ActionConfig.RESTClientGetter.ToRESTConfig()
	if err != nil {
		klog.V(1).Info(err, "failed to get rest config")
	}

	// [TODO] 나머지 테스트 완료 후 활성화
	token, err := utils.ReadDefaultToken()
	if err != nil {
		klog.V(1).Info(err, "failed to read default token")
	}

	cfg.BearerToken = token
	cfg.BearerTokenFile = ""

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
		klog.V(1).Info(err, "failed to set Env settings")
		return err
	}

	if err = hcm.Hcs.ActionConfig.Init(
		clientGetter,
		ns,
		os.Getenv("HELM_DRIVER"),
		hcm.Hcs.ActionConfig.Log,
	); err != nil {
		klog.V(1).Info(err, "failed to init action config")
		return err
	}

	hcm.Hcs.Providers = getter.All(settings)
	return nil
}

// func (hcm *HelmClientManager) SetClientTLS(serverName string) error {
// 	cfg, err := hcm.Hcs.ActionConfig.RESTClientGetter.ToRESTConfig()
// 	if err != nil {
// 		klog.Errorln(err, "failed to get rest config")
// 	}
// 	cfg.TLSClientConfig.CertFile = public_key
// 	cfg.TLSClientConfig.ServerName = serverName

// 	settings := hcm.Hcs.Settings
// 	options := &helmclient.Options{
// 		RepositoryCache:  repositoryCache,
// 		RepositoryConfig: repositoryConfig,
// 		Debug:            true,
// 		Linting:          true,
// 	}

// 	clientGetter := helmclient.NewRESTClientGetter("", nil, cfg)

// 	if err = internal.SetEnvSettings(options, settings); err != nil {
// 		klog.Errorln(err, "failed to set Env settings")
// 		return err

// 	}

// 	if err = hcm.Hcs.ActionConfig.Init(
// 		clientGetter,
// 		"",
// 		os.Getenv("HELM_DRIVER"),
// 		hcm.Hcs.ActionConfig.Log,
// 	); err != nil {
// 		klog.Errorln(err, "failed to init action config")
// 		return err
// 	}

// 	hcm.Hcs.Providers = getter.All(settings)
// 	return nil

// }
