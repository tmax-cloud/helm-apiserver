package schemas

import (
	"helm.sh/helm/v3/pkg/release"
)

type ReleaseRequest struct {
	ReleaseName string `json:"releasename,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

type ReleaseResponse struct {
	Release []release.Release `json:"release,omitempty"`
}
