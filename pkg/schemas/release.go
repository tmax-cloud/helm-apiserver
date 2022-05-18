package schemas

import (
	"helm.sh/helm/v3/pkg/chart"
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
	Release []Release `json:"release,omitempty"`
}

type Error struct {
	Error       string `json:"error,omitempty"`
	Description string `json:"description,omitempty"`
}

type Release struct {
	// Name is the name of the release
	Name string `json:"name,omitempty"`
	// Info provides information about a release
	Info *release.Info `json:"info,omitempty"`
	// Chart is the chart that was released.
	Chart *chart.Chart `json:"chart,omitempty"`
	// Config is the set of extra Values added to the chart.
	// These values override the default values inside of the chart.
	Config map[string]interface{} `json:"config,omitempty"`
	// Manifest is the string representation of the rendered template.
	Manifest string `json:"manifest,omitempty"`
	// Hooks are all of the hooks declared for this release.
	Hooks []*release.Hook `json:"hooks,omitempty"`
	// Version is an int which represents the revision of the release.
	Version int `json:"version,omitempty"`
	// Namespace is the kubernetes namespace of the release.
	Namespace string `json:"namespace,omitempty"`
	// Labels of the release.
	// Disabled encoding into Json cause labels are stored in storage driver metadata field.
	Labels map[string]string `json:"-"`

	// 직접 추가한 field
	Objects map[string]string `json:"objects,omitempty"`
}
