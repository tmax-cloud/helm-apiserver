package apis

import (
	"fmt"
	"net/http"

	"github.com/tmax-cloud/helm-apiserver/internal/apiserver"
	"github.com/tmax-cloud/helm-apiserver/internal/hclient"
	"github.com/tmax-cloud/helm-apiserver/internal/utils"
	v1 "github.com/tmax-cloud/helm-apiserver/pkg/apiserver/apis/v1"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tmax-cloud/helm-apiserver/internal/wrapper"
	"github.com/tmax-cloud/helm-apiserver/pkg/apiserver/apis/v1/chart"
)

type handler struct {
	v1Handler   apiserver.APIHandler
	helmHandler apiserver.APIHandler
}

// NewHandler instantiates a new apis handler
func NewHandler(parent wrapper.RouterWrapper, cli client.Client, hcm *hclient.HelmClientManager, authCli authorization.AuthorizationV1Interface, chartCache *chart.ChartCache, defaultRepo bool) (apiserver.APIHandler, error) {
	handler := &handler{}

	//apis
	apiWrapper := wrapper.New("/apis", nil, handler.apisHandler)
	if err := parent.Add(apiWrapper); err != nil {
		return nil, err
	}

	// /apis/v1
	v1Handler, err := v1.NewHandlerForAggr(apiWrapper, cli, hcm, authCli, chartCache, defaultRepo)
	if err != nil {
		return nil, err
	}
	handler.v1Handler = v1Handler

	//helm
	helmWrapper := wrapper.New("/helm", nil, nil)
	if err := parent.Add(helmWrapper); err != nil {
		return nil, err
	}

	// /helm/v1
	helmHandler, err := v1.NewHandlerForNormal(helmWrapper, cli, hcm, authCli, chartCache)
	if err != nil {
		return nil, err
	}
	handler.helmHandler = helmHandler

	return handler, nil
}

func (h *handler) apisHandler(w http.ResponseWriter, _ *http.Request) {
	groupVersion := metav1.GroupVersionForDiscovery{
		GroupVersion: fmt.Sprintf("%s/%s", apiserver.APIGroup, v1.APIVersion),
		Version:      v1.APIVersion,
	}

	group := metav1.APIGroup{}
	group.Kind = "APIGroup"
	group.Name = apiserver.APIGroup
	group.PreferredVersion = groupVersion
	group.Versions = append(group.Versions, groupVersion)

	apiGroupList := &metav1.APIGroupList{}
	apiGroupList.Kind = "APIGroupList"
	apiGroupList.Groups = append(apiGroupList.Groups, group)

	_ = utils.RespondJSON(w, apiGroupList)
}
