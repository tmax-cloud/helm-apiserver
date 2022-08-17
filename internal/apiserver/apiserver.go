package apiserver

import (
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"

	authorization "k8s.io/api/authorization/v1"
)

const (
	// APIGroup is a common group for api
	APIGroup   = "helmapi.tmax.io"
	APIVersion = "v1"

	// NamespaceParamKey is a common key for namespace var
	NamespaceParamKey = "namespace"

	userHeader   = "X-Remote-User"
	groupHeader  = "X-Remote-Group"
	extrasHeader = "X-Remote-Extra-"
)

var SchemaGroupVersion = schema.GroupVersion{Group: APIGroup, Version: APIVersion}

var (
	// SchemeBuilder is the scheme builder with scheme init functions to run for this API package
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme is a common registration function for mapping packaged scoped group & version keys to a scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemaGroupVersion)
	klog.Info("is group registered..???")
	klog.Info(scheme.IsGroupRegistered(APIGroup))
	return nil
}

// APIHandler is an api handler interface.
// Common functions should be defined, if needed
type APIHandler interface{}

// GetUserName extracts user name from the header
func GetUserName(header http.Header) (string, error) {
	for k, v := range header {
		if k == userHeader {
			return v[0], nil
		}
	}
	return "", fmt.Errorf("no header %s", userHeader)
}

// GetUserGroups extracts user group from the header
func GetUserGroups(header http.Header) ([]string, error) {
	for k, v := range header {
		if k == groupHeader {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no header %s", groupHeader)
}

// GetUserExtras extracts user extras from the header
func GetUserExtras(header http.Header) map[string]authorization.ExtraValue {
	extras := map[string]authorization.ExtraValue{}

	for k, v := range header {
		if strings.HasPrefix(k, extrasHeader) {
			extras[strings.TrimPrefix(k, extrasHeader)] = v
		}
	}

	return extras
}
