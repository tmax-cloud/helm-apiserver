package release

import (
	"fmt"
	"net/http"

	"github.com/tmax-cloud/helm-apiserver/internal/apiserver"
	"github.com/tmax-cloud/helm-apiserver/internal/hclient"
	"github.com/tmax-cloud/helm-apiserver/internal/wrapper"

	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

const (
	APIVersion      = "v1"
	repositoryCache = "/tmp/.helmcache"
)

type ReleaseHandler struct {
	hcm        *hclient.HelmClientManager
	authorizer apiserver.Authorizer
}

func NewHandler(parent wrapper.RouterWrapper, hcm *hclient.HelmClientManager, authCli authorization.AuthorizationV1Interface) (apiserver.APIHandler, error) {
	releaseHandler := &ReleaseHandler{hcm: hcm}

	// Authorizer
	releaseHandler.authorizer = apiserver.NewAuthorizer(authCli, apiserver.APIGroup, APIVersion, "update")

	// /v1/releases
	releaseAllWrapper := wrapper.New(fmt.Sprintf("/%s", "releases"), []string{http.MethodGet}, releaseHandler.releaseHandler)
	if err := parent.Add(releaseAllWrapper); err != nil {
		return nil, err
	}

	// /v1/releases/websocket
	releaseAllWebsocketWrapper := wrapper.New(fmt.Sprintf("/%s/%s", "releases", "websocket"), []string{http.MethodGet}, releaseHandler.Websocket)
	if err := parent.Add(releaseAllWebsocketWrapper); err != nil {
		return nil, err
	}

	// /v1/namespaces/<namespace>
	namespaceWrapper := wrapper.New("/namespaces/{ns-name}", nil, nil)
	if err := parent.Add(namespaceWrapper); err != nil {
		return nil, err
	}

	// /v1/namespaces/<namespace>/releases
	releaseWrapper := wrapper.New(fmt.Sprintf("/%s", "releases"), []string{http.MethodGet, http.MethodPost}, releaseHandler.releaseHandler)
	if err := namespaceWrapper.Add(releaseWrapper); err != nil {
		return nil, err
	}

	// /v1/namespaces/<namespace>/releases/websocket
	releaseWebsocketWrapper := wrapper.New(fmt.Sprintf("/%s/%s", "releases", "websocket"), []string{http.MethodGet}, releaseHandler.Websocket)
	if err := namespaceWrapper.Add(releaseWebsocketWrapper); err != nil {
		return nil, err
	}

	// /v1/namespaces/<namespace>/releases/<release-name>
	releaseParamWrapper := wrapper.New(fmt.Sprintf("/%s/{%s}", "releases", "release-name"), []string{http.MethodGet, http.MethodDelete, http.MethodPut}, releaseHandler.releaseHandler)
	if err := namespaceWrapper.Add(releaseParamWrapper); err != nil {
		return nil, err
	}
	releaseParamWrapper.Router().Use(releaseHandler.authorizer.Authorize)

	return releaseHandler, nil
}
