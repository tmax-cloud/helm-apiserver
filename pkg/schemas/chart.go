package schemas

import (
	"helm.sh/helm/v3/pkg/repo"
)

type ChartRequest struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	ChartRequestSpec `json:"chartrequestspec,omitempty"`
	Values           string `json:"values,omitempty"`
}

type ChartRequestSpec struct {
	PackageURL  string `json:"packageURL"`
	ReleaseName string `json:"releasename"`
	Version     string `json:"version"`
}

type ChartResponse struct {
	IndexFile repo.IndexFile         `json:"indexfile,omitempty"`
	Values    map[string]interface{} `json:"values,omitempty"` // UI 확인 필요
}
