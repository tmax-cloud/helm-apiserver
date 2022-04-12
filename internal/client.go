package internal

import (
	"os"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

const (
	repositoryCache  = "/tmp/.helmcache" // 캐시 디렉토리. 특정 chart-repo에 해당하는 chart 이름 리스트 txt파일과, 해당 repo의 index.yaml 파일이 저장됨
	repositoryConfig = "/tmp/.helmrepo"  // 현재 add된 repo들 저장.
)

func Client(options client.Options) (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	c, err := client.New(cfg, options)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewHelmClientInterface() helmclient.Client {
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Errorln(err, "failed to get rest config")
	}

	c, _ := client.New(cfg, client.Options{})

	sa, _ := GetServiceAccount(c, types.NamespacedName{Name: "helm-server-test-sa", Namespace: "helm-ns"})
	var secretName string

	for _, sec := range sa.Secrets {
		secretName = sec.Name
	}

	testSecret, _ := GetSecret(c, types.NamespacedName{Name: secretName, Namespace: "helm-ns"})
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

func NewHelmClientStruct() helmclient.HelmClient {
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

func newRESTClientGetter(namespace string, kubeConfig []byte, restConfig *rest.Config) *RESTClientGetter {
	return &RESTClientGetter{
		namespace:  namespace,
		kubeConfig: kubeConfig,
		restConfig: restConfig,
	}
}

func newClient(options *helmclient.Options, clientGetter genericclioptions.RESTClientGetter, settings *cli.EnvSettings) (helmclient.HelmClient, error) {
	setEnvSettings(options, settings)

	debugLog := options.DebugLog

	actionConfig := new(action.Configuration)
	actionConfig.Init(
		clientGetter,
		// settings.Namespace(), // all-namespace 구현을 위해 공백 처리
		"",
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
