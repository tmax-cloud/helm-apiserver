package schemas

import (
	"helm.sh/helm/v3/pkg/release"
)

type ReleaseRequest struct {
	// 일단은 path-variable로 구현해서 namespace 필요 없으나
	// GetRelease 제외 하고 필요할 수도 있으므로 일단 보류
	// Namespace          string `json:"namespace"`
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
