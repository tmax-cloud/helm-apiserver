package schemas

import (
	"helm.sh/helm/v3/pkg/repo"
)

type ChartResponse struct {
	IndexFile repo.IndexFile         `json:"indexfile,omitempty"`
	Values    map[string]interface{} `json:"values,omitempty"` // UI 확인 필요
}
