package schemas

import (
	"helm.sh/helm/v3/pkg/release"
)

type ReleaseRequest struct {
	ReleaseName string `json:"releasename"`
	Namespace   string `json:"namespace"`
}

type ReleaseResponse struct {
	Release release.Release `json:"release"`
}
