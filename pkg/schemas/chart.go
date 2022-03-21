package schemas

import (
	"time"

	"helm.sh/helm/v3/pkg/chart"
)

type ChartResponse struct {
	IndexFile IndexFile              `json:"indexfile,omitempty"`
	Values    map[string]interface{} `json:"values,omitempty"` // UI 확인 필요
}

type IndexFile struct {
	// This is used ONLY for validation against chartmuseum's index files and is discarded after validation.
	ServerInfo map[string]interface{}   `json:"serverInfo,omitempty"`
	APIVersion string                   `json:"apiVersion"`
	Generated  time.Time                `json:"generated"`
	Entries    map[string]ChartVersions `json:"entries"`
	PublicKeys []string                 `json:"publicKeys,omitempty"`

	// Annotations are additional mappings uninterpreted by Helm. They are made available for
	// other applications to add information to the index file.
	Annotations map[string]string `json:"annotations,omitempty"`
}

type ChartVersions []*ChartVersion

// ChartVersion represents a chart entry in the IndexFile
type ChartVersion struct {
	*chart.Metadata
	URLs    []string   `json:"urls"`
	Created time.Time  `json:"created,omitempty"`
	Removed bool       `json:"removed,omitempty"`
	Digest  string     `json:"digest,omitempty"`
	Repo    Repository `json:"repo,omitempty"`

	// ChecksumDeprecated is deprecated in Helm 3, and therefore ignored. Helm 3 replaced
	// this with Digest. However, with a strict YAML parser enabled, a field must be
	// present on the struct for backwards compatibility.
	ChecksumDeprecated string `json:"checksum,omitempty"`

	// EngineDeprecated is deprecated in Helm 3, and therefore ignored. However, with a strict
	// YAML parser enabled, this field must be present.
	EngineDeprecated string `json:"engine,omitempty"`

	// TillerVersionDeprecated is deprecated in Helm 3, and therefore ignored. However, with a strict
	// YAML parser enabled, this field must be present.
	TillerVersionDeprecated string `json:"tillerVersion,omitempty"`

	// URLDeprecated is deprecated in Helm 3, superseded by URLs. It is ignored. However,
	// with a strict YAML parser enabled, this must be present on the struct.
	URLDeprecated string `json:"url,omitempty"`
}
