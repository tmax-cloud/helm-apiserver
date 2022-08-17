package v1

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/tmax-cloud/helm-apiserver/internal/apiserver"
	"github.com/tmax-cloud/helm-apiserver/internal/hclient"
	"github.com/tmax-cloud/helm-apiserver/internal/utils"
	"github.com/tmax-cloud/helm-apiserver/internal/wrapper"
	"github.com/tmax-cloud/helm-apiserver/pkg/apiserver/apis/v1/chart"
	"github.com/tmax-cloud/helm-apiserver/pkg/apiserver/apis/v1/release"
	"github.com/tmax-cloud/helm-apiserver/pkg/apiserver/apis/v1/repos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// APIVersion of the apis
	APIVersion      = "v1"
	ChartResource   = "charts"
	ReleaseResource = "releases"
	RepoResource    = "repos"
)

type handler struct {
	chartHandler   apiserver.APIHandler
	releaseHandler apiserver.APIHandler
	repoHandler    apiserver.APIHandler
}

// NewHandler instantiates a new v1 api handler
func NewHandler(parent wrapper.RouterWrapper, cli client.Client, hcm *hclient.HelmClientManager, authCli authorization.AuthorizationV1Interface, logger logr.Logger, chartCache *chart.ChartCache) (apiserver.APIHandler, error) {
	handler := &handler{}

	// /v1
	versionWrapper := wrapper.New(fmt.Sprintf("/%s/%s", apiserver.APIGroup, APIVersion), nil, handler.versionHandler)
	if err := parent.Add(versionWrapper); err != nil {
		return nil, err
	}

	// /v1/charts/<chart>
	chartHandler, err := chart.NewHandler(versionWrapper, hcm, authCli, chartCache)
	if err != nil {
		return nil, err
	}
	handler.chartHandler = chartHandler

	// /v1/releases/<release>
	releaseHandler, err := release.NewHandler(versionWrapper, hcm, authCli)
	if err != nil {
		return nil, err
	}
	handler.releaseHandler = releaseHandler

	// /v1/repos/<repo>
	repoHandler, err := repos.NewHandler(versionWrapper, hcm, authCli, chartCache)
	if err != nil {
		return nil, err
	}
	handler.repoHandler = repoHandler

	return handler, nil
}

func (h *handler) versionHandler(w http.ResponseWriter, _ *http.Request) {
	apiResourceList := &metav1.APIResourceList{}
	apiResourceList.Kind = "APIResourceList"
	apiResourceList.GroupVersion = fmt.Sprintf("%s/%s", apiserver.APIGroup, APIVersion)
	apiResourceList.APIVersion = APIVersion

	apiResourceList.APIResources = []metav1.APIResource{
		{
			Name:       ChartResource,
			Namespaced: false,
		},
		{
			Name:       ReleaseResource,
			Namespaced: true, // release is namespaced scope
		},
		{
			Name:       RepoResource,
			Namespaced: false,
		},
	}

	_ = utils.RespondJSON(w, apiResourceList)
}
