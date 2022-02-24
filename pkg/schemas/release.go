package schemas

import (
	"helm.sh/helm/v3/pkg/release"
)

type ReleaseRequest struct {
	ReleaseRequestSpec `json:"releaseRequestSpec,omitempty"`
	Values             string `json:"values,omitempty"`
}

type ReleaseRequestSpec struct {
	PackageURL  string `json:"packageURL"`
	ReleaseName string `json:"releaseName"`
	Version     string `json:"version"`
}

type ReleaseResponse struct {
	Release []release.Release `json:"release,omitempty"`
}

type Error struct {
	Error       string `json:"error,omitempty"`
	Description string `json:"description,omitempty"`
}
