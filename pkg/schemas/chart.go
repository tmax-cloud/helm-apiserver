package schemas

import (
	"helm.sh/helm/v3/pkg/time"
)

type ChartRequest struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	ChartRequestSpec `json:"chartRequestSpec,omitempty"`
	Values           string `json:"values,omitempty"`
}

type ChartRequestSpec struct {
	PackageURL  string `json:"packageURL"`
	ReleaseName string `json:"releaseName"`
	Version     string `json:"version"`
}

/*
type ChartResponse struct {
	//IndexFile repo.IndexFile         `json:"indexfile,omitempty"`
	Values map[string]interface{} `json:"values,omitempty"` // UI 확인 필요
}
*/

type ChartsResponse struct {
	ChartInfos []ChartInfo `json:"chartInfos,omitempty"`
}

type IndexFile struct {
	APIVersion string                `json:"apiVersion"`
	Generated  time.Time             `json:"generated"`
	Entries    map[string]ChartInfos `json:"entries"`
	PublicKeys []string              `json:"publicKeys,omitempty"`
}

type ChartInfos []*ChartInfo

type ChartInfo struct {
	ChartMetadata
	URLs    []string  `json:"urls"`
	Created time.Time `json:"created,omitempty"`
	Removed bool      `json:"removed,omitempty"`
	Digest  string    `json:"digest,omitempty"`
}

type ChartMetadata struct {
	// The name of the chart
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The URL to a relevant project page, git repo, or contact person
	Home string `protobuf:"bytes,2,opt,name=home,proto3" json:"home,omitempty"`
	// Source is the URL to the source code of this chart
	Sources []string `protobuf:"bytes,3,rep,name=sources,proto3" json:"sources,omitempty"`
	// A SemVer 2 conformant version string of the chart
	Version string `protobuf:"bytes,4,opt,name=version,proto3" json:"version,omitempty"`
	// A one-sentence description of the chart
	Description string `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	// A list of string keywords
	Keywords []string `protobuf:"bytes,6,rep,name=keywords,proto3" json:"keywords,omitempty"`
	// A list of name and URL/email address combinations for the maintainer(s)
	//Maintainers []*Maintainer `protobuf:"bytes,7,rep,name=maintainers,proto3" json:"maintainers,omitempty"`
	// The name of the template engine to use. Defaults to 'gotpl'.
	Engine string `protobuf:"bytes,8,opt,name=engine,proto3" json:"engine,omitempty"`
	// The URL to an icon file.
	Icon string `protobuf:"bytes,9,opt,name=icon,proto3" json:"icon,omitempty"`
	// The API Version of this chart.
	ApiVersion string `protobuf:"bytes,10,opt,name=apiVersion,proto3" json:"apiVersion,omitempty"`
	// The condition to check to enable chart
	Condition string `protobuf:"bytes,11,opt,name=condition,proto3" json:"condition,omitempty"`
	// The tags to check to enable chart
	Tags string `protobuf:"bytes,12,opt,name=tags,proto3" json:"tags,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `protobuf:"bytes,13,opt,name=appVersion,proto3" json:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `protobuf:"varint,14,opt,name=deprecated,proto3" json:"deprecated,omitempty"`
	// TillerVersion is a SemVer constraints on what version of Tiller is required.
	// See SemVer ranges here: https://github.com/Masterminds/semver#basic-comparisons
	TillerVersion string `protobuf:"bytes,15,opt,name=tillerVersion,proto3" json:"tillerVersion,omitempty"`
	// Annotations are additional mappings uninterpreted by Tiller,
	// made available for inspection by other applications.
	Annotations map[string]string `` /* 164-byte string literal not displayed */
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion          string   `protobuf:"bytes,17,opt,name=kubeVersion,proto3" json:"kubeVersion,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}
