package chart

import (
	"fmt"
	"net/http"

	"github.com/tmax-cloud/helm-apiserver/internal/apiserver"
	"github.com/tmax-cloud/helm-apiserver/internal/hclient"
	"github.com/tmax-cloud/helm-apiserver/internal/wrapper"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	authorization "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

const (
	indexFileSuffix  = "-index.yaml"
	chartsFileSuffix = "-charts.txt"
	APIVersion       = "v1"
)

type ChartHandler struct {
	hcm *hclient.HelmClientManager
	*ChartCache
	authorizer apiserver.Authorizer
}

type ChartCache struct {
	Index              *schemas.IndexFile
	SingleChartEntries map[string]schemas.ChartVersions
}

func NewHandler(parent wrapper.RouterWrapper, hcm *hclient.HelmClientManager, authCli authorization.AuthorizationV1Interface, chartCache *ChartCache) (apiserver.APIHandler, error) {
	chartHandler := &ChartHandler{hcm: hcm, ChartCache: chartCache}

	// Authorizer
	chartHandler.authorizer = apiserver.NewAuthorizer(authCli, apiserver.APIGroup, APIVersion, "update")

	// /v1/charts
	chartWrapper := wrapper.New(fmt.Sprintf("/%s", "charts"), []string{http.MethodGet}, chartHandler.chartHandler)
	if err := parent.Add(chartWrapper); err != nil {
		return nil, err
	}
	// /v1/charts/<chart-name>
	chartParamWrapper := wrapper.New(fmt.Sprintf("/%s/{%s}", "charts", "chart-name"), []string{http.MethodGet}, chartHandler.chartHandler)
	if err := parent.Add(chartParamWrapper); err != nil {
		return nil, err
	}
	// /v1/charts/<chart-name>/versions/<version>
	chartVersionWrapper := wrapper.New(fmt.Sprintf("/%s/{%s}", "versions", "version"), []string{http.MethodGet}, chartHandler.chartHandler)
	if err := chartParamWrapper.Add(chartVersionWrapper); err != nil {
		return nil, err
	}

	// chartWrapper.Router().Use(chartHandler.authorizer.Authorize)

	return chartHandler, nil
}
