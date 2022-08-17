package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/tmax-cloud/helm-apiserver/internal/apiserver"
	"github.com/tmax-cloud/helm-apiserver/internal/hclient"
	"github.com/tmax-cloud/helm-apiserver/internal/utils"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tmax-cloud/helm-apiserver/internal/wrapper"
	"github.com/tmax-cloud/helm-apiserver/pkg/apiserver/apis"
	"github.com/tmax-cloud/helm-apiserver/pkg/apiserver/apis/v1/chart"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
)

var log = logf.Log.WithName("api-server")

// Server is an interface of server
type Server interface {
	Start()
}

// server is an api server
type server struct {
	wrapper     wrapper.RouterWrapper
	client      client.Client
	authCli     authorization.AuthorizationV1Interface
	cache       cache.Cache
	hcm         *hclient.HelmClientManager
	apisHandler apiserver.APIHandler

	*chart.ChartCache
}

// +kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,namespace=kube-system,resourceNames=extension-apiserver-authentication,verbs=get;list;watch

// New is a constructor of server
func New(cli client.Client, cfg *rest.Config, hcm *hclient.HelmClientManager, cache cache.Cache) (Server, error) {
	var err error
	srv := &server{}
	srv.wrapper = wrapper.New("/", nil, srv.rootHandler)
	srv.wrapper.SetRouter(mux.NewRouter())
	srv.wrapper.Router().HandleFunc("/", srv.rootHandler)
	srv.client = cli
	srv.cache = cache
	srv.hcm = hcm
	chartCache := &chart.ChartCache{
		Index: getIndex(),
	}
	srv.ChartCache = chartCache
	srv.ChartCache.SingleChartEntries = getSingleChart(srv.ChartCache.Index)
	srv.authCli, err = authorization.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Set apisHandler
	apisHandler, err := apis.NewHandler(srv.wrapper, srv.client, srv.hcm, srv.authCli, log, srv.ChartCache)
	if err != nil {
		return nil, err
	}
	srv.apisHandler = apisHandler

	return srv, nil
}

// Start starts the server
func (s *server) Start() {
	// Wait for the cache init
	if cacheInit := s.cache.WaitForCacheSync(context.Background()); !cacheInit {
		panic(fmt.Errorf("cannot wait for cache init"))
	}

	// Create cert
	if err := createCert(context.Background(), s.client); err != nil {
		panic(err)
	}

	addr := "0.0.0.0:8443"
	log.Info(fmt.Sprintf("API aggregation server is running on %s", addr))

	cfg, err := tlsConfig(context.Background(), s.client)
	if err != nil {
		panic(err)
	}
	klog.Info("serving https")
	httpServer := &http.Server{Addr: ":8443", Handler: s.wrapper.Router(), TLSConfig: cfg}
	if err := httpServer.ListenAndServeTLS(path.Join(certDir, "tls.crt"), path.Join(certDir, "tls.key")); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (s *server) rootHandler(w http.ResponseWriter, _ *http.Request) {
	paths := metav1.RootPaths{}

	addPath(&paths.Paths, s.wrapper)

	_ = utils.RespondJSON(w, paths)
}

// addPath adds all the leaf API endpoints
func addPath(paths *[]string, w wrapper.RouterWrapper) {
	if w.Handler() != nil {
		*paths = append(*paths, w.FullPath())
	}

	for _, c := range w.Children() {
		addPath(paths, c)
	}
}

func getIndex() *schemas.IndexFile {
	// Read repositoryConfig File which contains repo Info list
	repoList, err := utils.ReadRepoList()
	if err != nil {
		klog.Errorln(err, "failed to save index file")
		return nil
	}

	repoInfos := make(map[string]string)
	// store repo names
	for _, repoInfo := range repoList.Repositories {
		repoInfos[repoInfo.Name] = repoInfo.Url
	}

	index := &schemas.IndexFile{}
	allEntries := make(map[string]schemas.ChartVersions)
	// col := db.GetMongoDBConnetion() // #######테스트########

	// read all index.yaml file and save only Entries
	for repoName, repoUrl := range repoInfos {
		if index, err = utils.ReadRepoIndex(repoName); err != nil {
			klog.Errorln(err, "failed to read index file")
		}

		// add repo info
		for key, charts := range index.Entries {
			for _, chart := range charts {
				chart.Repo.Name = repoName
				chart.Repo.Url = repoUrl
				// _, err := db.InsertDoc(col, chart) // #######테스트########
				// klog.Info("insert done!")
				// if err != nil {
				// 	klog.Error(err)
				// }
			}
			allEntries[repoName+"_"+key] = charts // 중복 chart name 가능하도록 repo name과 결합
		}
	}

	// filter := bson.D{{}}
	// var test []schemas.ChartVersion
	// test, _ = db.FindDoc(col, filter, filter)
	// for _, ch := range test {
	// 	klog.Info(ch.Name)
	// }

	index.Entries = allEntries
	klog.Info("saving index file is done")
	return index
}

func getSingleChart(index *schemas.IndexFile) map[string]schemas.ChartVersions {
	if index == nil {
		return nil
	}

	singleChartEntries := make(map[string]schemas.ChartVersions)
	for key, charts := range index.Entries {
		var oneChart []*schemas.ChartVersion
		oneChart = append(oneChart, charts[0])
		singleChartEntries[key] = oneChart
	}
	return singleChartEntries
}
